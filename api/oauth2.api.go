package api

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/fido-alliance/fdo-fido-conformance-server/api/commonapi"
	"github.com/fido-alliance/fdo-fido-conformance-server/dbs"
	"github.com/fido-alliance/fdo-fido-conformance-server/services"
	fdoshared "github.com/fido-alliance/fdo-shared"
	"github.com/gorilla/mux"
)

type OAuth2API struct {
	UserDB        *dbs.UserTestDB
	SessionDB     *dbs.SessionDB
	OAuth2Service *services.OAuth2Service
	Notify        services.NotifyService
}

func (h *OAuth2API) checkAutzAndGetSession(r *http.Request) (*dbs.SessionEntry, error) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return nil, errors.New("Failed to read cookie. " + err.Error())

	}

	if sessionCookie == nil {
		return nil, errors.New("cookie does not exists")
	}

	sessionInst, err := h.SessionDB.GetSessionEntry([]byte(sessionCookie.Value))
	if err != nil {
		return nil, errors.New("Session expired. " + err.Error())
	}

	return sessionInst, nil
}

func (h *OAuth2API) setUserSession(w http.ResponseWriter, sessionInst dbs.SessionEntry) error {
	sessionDbId, err := h.SessionDB.NewSessionEntry(sessionInst)
	if err != nil {
		return errors.New("Error creating session. " + err.Error())
	}

	http.SetCookie(w, commonapi.GenerateCookie(sessionDbId))
	return nil
}

type OAuth2RedictUrlResult struct {
	RedirectUrl string                     `json:"redirect_url"`
	Status      commonapi.FdoConfApiStatus `json:"status"`
}

func (h *OAuth2API) InitWithRedirectUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		commonapi.RespondError(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	if r.Context().Value(fdoshared.CFG_ENV_MODE) == fdoshared.CFG_MODE_ONPREM {
		log.Println("Only allowed for on-line build!")
		commonapi.RespondError(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	providerid := services.OAuth2ProviderID(vars["providerid"])

	oauth2Provider, err := h.OAuth2Service.GetProvider(providerid)
	if err != nil {
		commonapi.RespondError(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	redirectUrl, state, nonce := oauth2Provider.GetRedirectUrl()

	err = h.setUserSession(w, dbs.SessionEntry{
		OAuth2Provider: string(providerid),
		OAuth2Nonce:    nonce,
		OAuth2State:    state,
	})
	if err != nil {
		log.Println("Error creating session. " + err.Error())
		commonapi.RespondError(w, "Internal server error. ", http.StatusBadRequest)
		return
	}

	commonapi.RespondSuccessStruct(w, OAuth2RedictUrlResult{RedirectUrl: redirectUrl, Status: commonapi.FdoApiStatus_OK})
}

func (h *OAuth2API) ProcessCallback(w http.ResponseWriter, r *http.Request) {
	session, err := h.checkAutzAndGetSession(r)
	if err != nil {
		log.Println(err)
		commonapi.RespondError(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	if r.Context().Value(fdoshared.CFG_ENV_MODE) == fdoshared.CFG_MODE_ONPREM {
		log.Println("Only allowed for on-line build!")
		commonapi.RespondError(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		commonapi.RespondError(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	providerid := services.OAuth2ProviderID(vars["providerid"])

	if providerid != services.OAuth2ProviderID(session.OAuth2Provider) || state != session.OAuth2State {
		commonapi.RespondError(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	oauth2Provider, err := h.OAuth2Service.GetProvider(providerid)
	if err != nil {
		commonapi.RespondError(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	email, isFidoGithubMember, err := oauth2Provider.GetUserInfo(code)
	if err != nil {
		log.Println(err)
		commonapi.RespondError(w, "Failed to validate OAuth2 code!", http.StatusUnauthorized)
		return
	}

	userInst, err := h.UserDB.Get(email)

	// User exists
	if err == nil && userInst != nil {
		if isFidoGithubMember || userInst.Status == dbs.AS_Validated {
			err = h.setUserSession(w, dbs.SessionEntry{
				Email:    email,
				LoggedIn: true,
			})
			if err != nil {
				log.Println("Error creating session. " + err.Error())
				commonapi.RespondError(w, "Internal server error. ", http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, commonapi.REDIRECT_HOME, http.StatusTemporaryRedirect)
			return
		} else {
			err = h.setUserSession(w, dbs.SessionEntry{
				Email: email,
			})
			if err != nil {
				log.Println("Error creating session. " + err.Error())
				commonapi.RespondError(w, "Internal server error. ", http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, commonapi.REDIRECT_AWAITING_VERIFICATION, http.StatusTemporaryRedirect)
			return
		}

	} else { // New user
		if !isFidoGithubMember {
			err = h.setUserSession(w, dbs.SessionEntry{
				Email:                strings.ToLower(email),
				OAuth2Email:          strings.ToLower(email),
				OAuth2AdditionalInfo: true,
			})
			if err != nil {
				log.Println("Error creating session. " + err.Error())
				commonapi.RespondError(w, "Internal server error. ", http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, commonapi.REDIRECT_ADDITIONAL_INFO, http.StatusTemporaryRedirect)
			return
		}

		err = h.UserDB.Save(dbs.UserTestDBEntry{
			Email:         strings.ToLower(email),
			Status:        dbs.AS_Validated,
			EmailVerified: true,
		})
		if err != nil {
			log.Println("Error saving user. " + err.Error())
			commonapi.RespondError(w, "Internal server error.", http.StatusInternalServerError)
			return
		}

		if isFidoGithubMember {
			err = h.setUserSession(w, dbs.SessionEntry{
				Email:    email,
				LoggedIn: true,
			})
			if err != nil {
				log.Println("Error creating session. " + err.Error())
				commonapi.RespondError(w, "Internal server error. ", http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, commonapi.REDIRECT_HOME, http.StatusTemporaryRedirect)
		} else {
			err = h.setUserSession(w, dbs.SessionEntry{
				Email:                strings.ToLower(email),
				OAuth2Email:          strings.ToLower(email),
				OAuth2AdditionalInfo: true,
			})
			if err != nil {
				log.Println("Error creating session. " + err.Error())
				commonapi.RespondError(w, "Internal server error. ", http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, commonapi.REDIRECT_AWAITING_VERIFICATION, http.StatusTemporaryRedirect)
		}
	}
}
