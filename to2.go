package main

// import (
// 	"io/ioutil"
// 	"log"
// 	"net/http"

// 	"github.com/WebauthnWorks/fdo-do/fdoshared"
// 	"github.com/fxamacker/cbor/v2"
// )

// const agreedWaitSeconds uint32 = 30 * 24 * 60 * 60 // 1 month

// type DoTo2 struct {
// 	session       *SessionDB
// 	HelloDeviceDB *HelloDeviceDB
// }

// // func (h *DoTo2) HelloDevice60(w http.ResponseWriter, r *http.Request) {
// // 	log.Println("Receiving HelloDevice60...")
// // 	if !CheckHeaders(w, r, fdoshared.TO0_HELLO_20) {
// // 		return
// // 	}

// // bodyBytes, err := ioutil.ReadAll(r.Body)
// // if err != nil {
// // 	RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_HELLO_DEVICE_60, "Failed to read body!", http.StatusBadRequest)
// // 	return
// // }

// // 	var helloDevice fdoshared.HelloDevice60
// // 	err = cbor.Unmarshal(bodyBytes, &helloDevice)
// // 	if err != nil {
// // 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_HELLO_DEVICE_60, "Failed to decode body!", http.StatusBadRequest)
// // 		return
// // 	}

// // 	// 1. Obtain voucher
// // 	// 1a. Marshal OVHeader from voucher
// // 	var voucher fdoshared.OwnershipVoucher
// // 	OVHeaderBytes, _ := cbor.Marshal(voucher.OVHeader)

// // 	// 2. Begin Key Exchange
// // 	// Write code here.
// // 	// xAKeyExchange, err := beginECDHKeyExchange(kexSuiteName)

// // 	// 3. Generate Nonce
// // 	NonceTO2ProveOV := make([]byte, 16)
// // 	rand.Read(NonceTO2ProveOV)

// // 	// 4. Encode response

// // 	err = h.HelloDeviceDB.Save(helloDevice.Guid, helloDevice, agreedWaitSeconds)

// // 	// Response:

// // 	newSessionInst := SessionEntry{
// // 		Protocol:        fdoshared.To2,
// // 		NonceTO2ProveOV: helloDevice.NonceTO2ProveOV,
// // 	}

// // 	sessionId, err := h.session.NewSessionEntry(newSessionInst)
// // 	if err != nil {
// // 		RespondFDOError(w, r, fdoshared.INTERNAL_SERVER_ERROR, fdoshared.TO2_HELLO_DEVICE_60, "Internal Server Error!", http.StatusInternalServerError)
// // 		return
// // 	}

// // 	helloDeviceHash, err := fdoshared.GenerateFdoHash(bodyBytes, -16) // fix
// // 	if err != nil {
// // 		RespondFDOError(w, r, fdoshared.INTERNAL_SERVER_ERROR, fdoshared.TO2_HELLO_DEVICE_60, "Internal Server Error!", http.StatusInternalServerError)
// // 		return
// // 	}

// // NonceTO2ProveDv61
// // store NonceTO2ProveDv61

// // 	proveOVHdrPayload := fdoshared.TO2ProveOVHdrPayload{
// // 		OVHeader:            OVHeaderBytes,
// // 		NumOVEntries:        255,                    // change
// // 		HMac:                fdoshared.HashOrHmac{}, // Ownership Voucher "hmac" of hdr
// // 		NonceTO2ProveOV:     helloDevice.NonceTO2ProveOV,
// // 		EBSigInfo:           helloDevice.EASigInfo,
// // 		XAKeyExchange:       "string", // Key exchange first step
// // 		HelloDeviceHash:     helloDeviceHash,
// // 		MaxOwnerMessageSize: helloDevice.MaxDeviceMessageSize, // change
// // 	}

// // 	proveOVHdrPayloadBytes, _ := cbor.Marshal(proveOVHdrPayload)

// // 	helloAck, _ := fdoshared.GenerateCoseSignature(proveOVHdrPayloadBytes, fdoshared.ProtectedHeader{}, fdoshared.UnprotectedHeader{}, mfgPrivateKey, sgType)
// // 	// fdoshared.ProveOVHdr61

// // 	helloAckBytes, _ := cbor.Marshal(helloAck)

// // 	sessionIdToken := "Bearer " + string(sessionId)
// // 	w.Header().Set("Authorization", sessionIdToken)
// // 	w.Header().Set("Content-Type", fdoshared.CONTENT_TYPE_CBOR)
// // 	w.Header().Set("Message-Type", fdoshared.TO0_HELLO_ACK_21.ToString())
// // 	w.WriteHeader(http.StatusOK)
// // 	w.Write(helloAckBytes)
// // }

// func (h *DoTo2) GetOVNextEntry62(w http.ResponseWriter, r *http.Request) {

// 	if !CheckHeaders(w, r, fdoshared.TO0_OWNER_SIGN_22) {
// 		return
// 	}

// 	headerIsOk, sessionId, _ := ExtractAuthorizationHeader(w, r, fdoshared.TO0_OWNER_SIGN_22)
// 	if !headerIsOk {
// 		return
// 	}

// 	session, err := h.session.GetSessionEntry(sessionId)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_GET_OVNEXTENTRY_62, "Unauthorized (1)", http.StatusUnauthorized)
// 		return
// 	}

// 	bodyBytes, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_GET_OVNEXTENTRY_62, "Failed to read body!", http.StatusBadRequest)
// 		return
// 	}

// 	voucher := session.Voucher

// 	var getOVNextEntry fdoshared.GetOVNextEntry62
// 	err = cbor.Unmarshal(bodyBytes, &getOVNextEntry)

// 	// check to see if LastOVEntryNum was never set, if so then the OVEntryNum must call 0
// 	if session.LastOVEntryNum == 0 && getOVNextEntry.OVEntryNum != 0 {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_GET_OVNEXTENTRY_62, "2 Error with OVEntryNum!", http.StatusBadRequest)
// 		return
// 	}

// 	if getOVNextEntry.OVEntryNum != session.LastOVEntryNum+1 {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_GET_OVNEXTENTRY_62, "3 Error with OVEntryNum!", http.StatusBadRequest)
// 		return
// 	}

// 	// update OVEntryNum in session storage
// 	session.LastOVEntryNum = getOVNextEntry.OVEntryNum
// 	h.session.UpdateSessionEntry(sessionId, *session)

// 	if getOVNextEntry.OVEntryNum == session.TO2ProveOVHdrPayload.NumOVEntries-1 {
// 		// nextState = TO2.ProveDevice.
// 	} else {
// 		// nextState = getOVNextEntry
// 	}

// 	OVEntry := voucher.OVEntryArray[getOVNextEntry.OVEntryNum]

// 	var ovNextEntry63 = fdoshared.OVNextEntry63{
// 		OVEntryNum: getOVNextEntry.OVEntryNum,
// 		OVEntry:    OVEntry,
// 	}

// 	ovNextEntryBytes, _ := cbor.Marshal(ovNextEntry63)

// 	sessionIdToken := "Bearer " + string(sessionId)
// 	w.Header().Set("Authorization", sessionIdToken)
// 	w.Header().Set("Content-Type", fdoshared.CONTENT_TYPE_CBOR)
// 	w.Header().Set("Message-Type", fdoshared.TO2_OV_NEXTENTRY_63.ToString())
// 	w.WriteHeader(http.StatusOK)
// 	w.Write(ovNextEntryBytes)
// }

// func (h *DoTo2) ProveDevice64(w http.ResponseWriter, r *http.Request) {
// 	log.Println("Receiving ProveDevice64...")

// 	if !CheckHeaders(w, r, fdoshared.TO2_PROVE_DEVICE_64) {
// 		return
// 	}

// 	headerIsOk, sessionId, _ := ExtractAuthorizationHeader(w, r, fdoshared.TO2_PROVE_DEVICE_64)
// 	if !headerIsOk {
// 		return
// 	}

// 	session, err := h.session.GetSessionEntry(sessionId)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_PROVE_DEVICE_64, "Unauthorized (1)", http.StatusUnauthorized)
// 		return
// 	}

// 	bodyBytes, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_PROVE_DEVICE_64, "Failed to read body!", http.StatusBadRequest)
// 		return
// 	}

// 	voucher := session.Voucher

// 	var proveDevice64 fdoshared.ProveDevice64
// 	err = cbor.Unmarshal(bodyBytes, &proveDevice64)

// 	voucher.OVDevCertChain
// 	var placeHolder_publicKey fdoshared.FdoPublicKey
// 	signatureIsValid, err := fdoshared.VerifyCoseSignature(proveDevice64, placeHolder_publicKey)
// 	if err != nil {
// 		log.Println("ProveDevice64: Error verigetInfo_response[GetInfoRespKeys.fying. " + err.Error())
// 		RespondFDOError(w, r, fdoshared.INVALID_MESSAGE_ERROR, fdoshared.TO2_PROVE_DEVICE_64, "Failed to verify signature ProveToRV32, some error", http.StatusBadRequest)
// 		return
// 	}

// 	if !signatureIsValid {
// 		log.Println("ProveDevice64: Signature is not valid!")
// 		RespondFDOError(w, r, fdoshared.INVALID_MESSAGE_ERROR, fdoshared.TO2_PROVE_DEVICE_64, "Failed to verify signature!", http.StatusBadRequest)
// 		return
// 	}

// 	var EATPayloadBase fdoshared.EATPayloadBase
// 	err = cbor.Unmarshal(proveDevice64.Payload, &EATPayloadBase)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_PROVE_DEVICE_64, "Failed to decode body!", http.StatusBadRequest)
// 		return
// 	}

// 	NonceTO2ProveDv := EATPayloadBase.EatNonce
// 	TO2ProveDevicePayload := EATPayloadBase.EatFDO.TO2ProveDevicePayload
// 	NonceTO2SetupDv := proveDevice64.Unprotected.CUPHNonce

// 	// Complete Key Exchange here

// 	// TODO:
// 	TO2SetupDevicePayload := fdoshared.TO2SetupDevicePayload {
// 		RendezvousInfo: []fdoshared.RendezvousInstrList{},
// 		Guid: fdoshared.FdoGuid{},
// 		NonceTO2SetupDv: NonceTO2SetupDv,
// 		Owner2Key: nil,
// 	}

// 	var TO2SetupDevicePayloadBytes []byte
// 	cbor.Marshal([]byte, &TO2SetupDevicePayloadBytes)

// 	var SetupDevice65 = fdoshared.SetupDevice65{
// 		Protected: proveDevice64.Protected,
// 		Unprotected: proveDevice64.Unprotected,
// 		Payload: TO2SetupDevicePayloadBytes,
// 		Signature: nil,
// 	}

// 	SetupDeviceBytes, _ := cbor.Marshal(SetupDevice65)

// 	sessionIdToken := "Bearer " + string(sessionId)
// 	w.Header().Set("Authorization", sessionIdToken)
// 	w.Header().Set("Content-Type", fdoshared.CONTENT_TYPE_CBOR)
// 	w.Header().Set("Message-Type", fdoshared.TO2_OV_NEXTENTRY_63.ToString())
// 	w.WriteHeader(http.StatusOK)
// 	w.Write(SetupDeviceBytes)

// }

// func (h *DoTo2) DeviceServiceInfoReady66(w http.ResponseWriter, r *http.Request) {
// 	log.Println("Receiving Done70...")

// 	if !CheckHeaders(w, r, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66) {
// 		return
// 	}

// 	headerIsOk, sessionId, _ := ExtractAuthorizationHeader(w, r, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66)
// 	if !headerIsOk {
// 		return
// 	}

// 	session, err := h.session.GetSessionEntry(sessionId)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66, "Unauthorized (1)", http.StatusUnauthorized)
// 		return
// 	}

// 	bodyBytes, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DEVICE_SERVICE_INFO_READY_66, "Failed to read body!", http.StatusBadRequest)
// 		return
// 	}
// 	// bodyBytes will be encrypted
// 	// need to decrypt it using the sessionKey

// 	// var DeviceServiceInfo68 fdoshared.DeviceServiceInfo68
// 	// err = cbor.Unmarshal(bodyBytes, &DeviceServiceInfo68)

// }

// // // func (h *DoTo2) DeviceServiceInfo68() (*fdoshared.OwnerServiceInfo69, error) {
// // // 	return nil, nil
// // // }

// func (h *DoTo2) Done70(w http.ResponseWriter, r *http.Request) {
// 	log.Println("Receiving Done70...")

// 	if !CheckHeaders(w, r, fdoshared.TO2_DONE_70) {
// 		return
// 	}

// 	headerIsOk, sessionId, _ := ExtractAuthorizationHeader(w, r, fdoshared.TO2_DONE_70)
// 	if !headerIsOk {
// 		return
// 	}

// 	session, err := h.session.GetSessionEntry(sessionId)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DONE_70, "Unauthorized (1)", http.StatusUnauthorized)
// 		return
// 	}

// 	bodyBytes, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DONE_70, "Failed to read body!", http.StatusBadRequest)
// 		return
// 	}

// 	var Done fdoshared.Done70
// 	err = cbor.Unmarshal(bodyBytes, &Done)

// 	// check to see Nonce is equal to the nonce that was sent in 61
// 	// Bytes compare..

// 	session, err := h.session.GetSessionEntry(sessionId)
// 	NonceTO2ProveDv61 := session.NonceTO2ProveDv61
// 	if bytes.Compare(NonceTO2ProveDv61, Done.NonceTO2ProveDv) != 0 {
// 		RespondFDOError(w, r, fdoshared.MESSAGE_BODY_ERROR, fdoshared.TO2_DONE_70, "Nonces did not match", http.StatusBadRequest)
// 		return
// 	}

// }

// }

// // // /**
// // // /60
// // // 1. Generate voucher
// // // 2. Begin Key Exchange
// // // 3. Generate Nonce
// // // 4. Encode response

// // // + stores items in db, set headers etc, generate auth token etc

// // // /62
// // // 1. Check previous entry, make sure this request is one entry higher
// // // 2.
// // // 3.
// // // 4.

// // // /64
// // // 1. Validate nonce is same as in 61
// // // 2. Complete exchange
// // // 3. Encode response

// // // /66
// // // 1. Decrypt message
// // // 2.
// // // 3.

// // // /68
// // // 0. Decrypt message
// // // 1. handleMaxDeviceServiceInfoSize
// // // 2. handleCheckDevModKeys
// // // 3. Encode response

// // // /70
// // // 0. Decrypt message
// // // 1. Get NonceTO2SetupDv from db
// // // 2. validateNonceDV (/70 = 61)
// // // 3. Encode response

// // // **/
