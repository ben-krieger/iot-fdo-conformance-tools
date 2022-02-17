package main

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/WebauthnWorks/fdo-do/fdoshared"
	"github.com/fxamacker/cbor/v2"
)

const MAX_DEVICE_SERVICE_INFO_SIZE = 1300

func (h *DoTo2) DeviceServiceInfoReady66(w http.ResponseWriter, r *http.Request) {
	log.Println("Receiving DeviceServiceInfoReady66...")

	if !CheckHeaders(w, r, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66) {

		RespondFDOError(w, r, fdoshared.INVALID_MESSAGE_ERROR, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66, "Failed to read body!", http.StatusBadRequest)
		return
	}

	headerIsOk, sessionId, _ := ExtractAuthorizationHeader(w, r, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66)
	if !headerIsOk {

		RespondFDOError(w, r, fdoshared.INVALID_MESSAGE_ERROR, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66, "Failed to decode body", http.StatusBadRequest)
		return
	}

	session, err := h.session.GetSessionEntry(sessionId)
	if err != nil {
		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66, "Unauthorized (1)", http.StatusUnauthorized)
		return
	}

	bodyBytes2, err := ioutil.ReadAll(r.Body)
	if err != nil {
		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66, "Failed to read body!", http.StatusBadRequest)
		return
	}

	// DELETE
	hex.EncodeToString(bodyBytes2)
	bodyBytesAsString := string(bodyBytes2)
	bodyBytes, err := hex.DecodeString(bodyBytesAsString)
	// DELETE

	sessionKey := session.SessionKey
	log.Println(sessionKey) // used to decrypt
	// Insert decryption logic here...
	decryptionBytes := bodyBytes
	// voucher := session.Voucher

	var DeviceServiceInfoReady66 fdoshared.DeviceServiceInfoReady66
	err = cbor.Unmarshal(decryptionBytes, &DeviceServiceInfoReady66)
	if err != nil {
		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66, "Failed to decode body!", http.StatusBadRequest)
		return
	}
	// bodyBytes will be encrypted
	// need to decrypt it using the sessionKey
	maxDeviceServiceInfoSz := DeviceServiceInfoReady66.MaxOwnerServiceInfoSz
	if maxDeviceServiceInfoSz == 0 {
		maxDeviceServiceInfoSz = MAX_DEVICE_SERVICE_INFO_SIZE
	}

	var OwnerServiceInfoReady = fdoshared.OwnerServiceInfoReady67{
		MaxDeviceServiceInfoSz: maxDeviceServiceInfoSz,
	}
	OwnerServiceInfoReadyBytes, err := cbor.Marshal(OwnerServiceInfoReady)
	if err != nil {
		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66, "Failed to decode body!", http.StatusBadRequest)
		return
	}
	// Encode(OwnerServiceInfoReadyBytes)
	// => Encrypt OwnerServiceInfoReadyBytes inside an ETM object

	sessionIdToken := "Bearer " + string(sessionId)
	w.Header().Set("Authorization", sessionIdToken)
	w.Header().Set("Content-Type", fdoshared.CONTENT_TYPE_CBOR)
	w.Header().Set("Message-Type", fdoshared.TO2_OWNER_SERVICE_INFO_READY_67.ToString())
	w.WriteHeader(http.StatusOK)
	w.Write(OwnerServiceInfoReadyBytes)
}
