package fdoshared

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"errors"
	"fmt"
	"math/big"

	"github.com/fxamacker/cbor/v2"
)

type UnprotectedHeader struct {
	CUPHNonce       []byte       `cbor:"256,keyasint,omitempty"`
	CUPHOwnerPubKey FdoPublicKey `cbor:"257,keyasint,omitempty"`
	EATMAROEPrefix  []byte       `cbor:"-258,keyasint,omitempty"`
	EUPHNonce       []byte       `cbor:"-259,keyasint,omitempty"`
	AESIV           []byte       `cbor:"5,keyasint,omitempty"`
}

type ProtectedHeader struct {
	Alg       int    `cbor:"1,keyasint,omitempty"`
	Kid       []byte `cbor:"4,keyasint,omitempty"`
	IV        []byte `cbor:"5,keyasint,omitempty"`
	PartialIV []byte `cbor:"6,keyasint,omitempty"`
}

type CoseSignature struct {
	_           struct{} `cbor:",toarray"`
	Protected   []byte
	Unprotected UnprotectedHeader
	Payload     []byte
	Signature   []byte
}

type EATPayloadBase struct {
	// EatFDO   []byte   `cbor:"-257,keyasint,omitempty"` // TODO change TYPE??
	// EatFDO   `cbor:"-257,keyasint,omitempty"` // TODO change TYPE??
	EatNonce []byte      `cbor:"10,keyasint,omitempty"`
	EatUEID  [17]byte    `cbor:"11,keyasint,omitempty"`
	EatFDO   EATPayloads `cbor:"-257,keyasint,omitempty"`
}

type EATPayloads struct {
	TO2ProveDevicePayload *TO2ProveDevicePayload
}

type TO2ProveDevicePayload struct {
	xBKeyExchange xBKeyExchange
}

type xAKeyExchange []byte
type xBKeyExchange []byte

type CoseContext string

const Signature1 CoseContext = "Signature1"

// Signature must be computed over a sig_structure:
// Sig_structure = [
//   context : "Signature" / "Signature1" / "CounterSignature",
//   body_protected : empty_or_serialized_map,
//   external_aad : bstr,
//   payload : bstr
// ]
//
// See the COSE RFC 8152 for details on this.

type CoseSignatureStructure struct {
	_           struct{} `cbor:",toarray"`
	Context     CoseContext
	Protected   []byte // Protected header. Serialized
	ExternalAAD []byte // External authentication data. For FDO keep it zero length byte string
	Payload     []byte
}

func NewSig1Payload(protectedHeader []byte, payload []byte) ([]byte, error) {
	sig1Inst := CoseSignatureStructure{
		Context:     Signature1,
		Protected:   protectedHeader,
		ExternalAAD: []byte{},
		Payload:     payload,
	}

	sig1Bytes, err := cbor.Marshal(sig1Inst)
	if err != nil {
		return []byte{}, errors.New("Error marshaling cose signature structure for Sig1. " + err.Error())
	}

	return sig1Bytes, nil
}

type FdoPkType uint8

const (
	RSA2048RESTR FdoPkType = 1  // RSA 2048 with restricted key/exponent (PKCS1 1.5 encoding)
	RSAPKCS      FdoPkType = 5  // RSA key, PKCS1, v1.5
	RSAPSS       FdoPkType = 6  // RSA key, PSS
	SECP256R1    FdoPkType = 10 // ECDSA secp256r1 = NIST-P-256 = prime256v1
	SECP384R1    FdoPkType = 11 // ECDSA secp384r1 = NIST-P-384
)

const (
	SECP256R1_SIG_LEN int = 64
	SECP384R1_SIG_LEN int = 96
)

type FdoPkEnc uint8

const (
	Crypto  FdoPkEnc = 0
	X509    FdoPkEnc = 1
	X5CHAIN FdoPkEnc = 2
	COSEKEY FdoPkEnc = 3
)

type FdoPublicKey struct {
	_      struct{} `cbor:",toarray"`
	PkType FdoPkType
	PkEnc  FdoPkEnc
	PkBody interface{}
}

type DeviceSgType int

const (
	StSECP256R1 DeviceSgType = -7
	StSECP384R1 DeviceSgType = -35
	StRSA2048   DeviceSgType = -257
	StRSA3072   DeviceSgType = -258
	StEPID10    DeviceSgType = 90
	StEPID11    DeviceSgType = 91
)

type SigInfo struct {
	_      struct{} `cbor:",toarray"`
	SgType DeviceSgType
	Info   string
}

type CoseConsts int

const (
	CoseOKP           CoseConsts = 1
	CoseEC2           CoseConsts = 2
	CoseRSA           CoseConsts = 3
	CoseECDHESHKDF256 CoseConsts = -25
	CoseES256         CoseConsts = -7
)

type CosePublicKey struct {
	Kty    CoseConsts  `cbor:"1,keyasint"`
	Alg    CoseConsts  `cbor:"3,keyasint"`
	CrvOrN interface{} `cbor:"-1,keyasint,omitempty"` // Could be Curve(EC/ED) int or N(RSA) []byte
	XorE   []byte      `cbor:"-2,keyasint,omitempty"` // Could be X(EC/ED) []byte or E(RSA) []byte
	Y      []byte      `cbor:"-3,keyasint,omitempty"`
}

func VerifyCertificateChain(chain [][]byte) (bool, []*x509.Certificate, error) {
	var finalChain []*x509.Certificate

	if len(chain) < 2 {
		return false, finalChain, errors.New("Failed to verify certificate chain. The length must be at least two!")
	}

	leafCertBytes := chain[0]
	leafCert, err := x509.ParseCertificate(leafCertBytes)
	if err != nil {
		return false, finalChain, errors.New("Error decoding leaf certificate. " + err.Error())

	}

	rootCertBytes := chain[len(chain)-1]
	rootCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		return false, finalChain, errors.New("Error decoding root certificate. " + err.Error())
	}

	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCert)

	interPool := x509.NewCertPool()

	if len(chain) > 2 {
		for i, interCertBytes := range chain[1 : len(chain)-1] {
			interCert, err := x509.ParseCertificate(interCertBytes)
			if err != nil {
				return false, finalChain, fmt.Errorf("Error decoding intermediate %d certificate. %s", i, err.Error())
			}
			interPool.AddCert(interCert)
		}
	}

	verificationChain, err := leafCert.Verify(x509.VerifyOptions{
		Intermediates: interPool,
		Roots:         rootPool,
	})
	if err != nil {
		return false, nil, errors.New("Error verifying certificate chain! " + err.Error())
	}
	finalChain = verificationChain[0]

	return true, finalChain, nil
}

func VerifyCoseSignatureWithCertificate(coseSig CoseSignature, publicKey FdoPublicKey, certs [][]byte) (bool, error) {
	newPubKey := FdoPublicKey{
		PkType: publicKey.PkType,
		PkEnc:  X5CHAIN,
		PkBody: certs,
	}

	return VerifyCoseSignature(coseSig, newPubKey)
}
func VerifySignature(payload []byte, signature []byte, publicKeyInst interface{}, pkType FdoPkType) (bool, error) {
	switch pkType {
	case SECP256R1:
		if len(signature) != SECP256R1_SIG_LEN {
			return false, errors.New("For ES256, signature must be 64 bytes long!")
		}

		payloadHash := sha256.Sum256(payload)

		r := new(big.Int)
		r.SetBytes(signature[0:32])

		s := new(big.Int)
		s.SetBytes(signature[32:64])

		return ecdsa.Verify(publicKeyInst.(*ecdsa.PublicKey), payloadHash[:], r, s), nil
	case SECP384R1:
		if len(signature) != SECP384R1_SIG_LEN {
			return false, errors.New("For ES384, signature must be 96 bytes long!")
		}

		payloadHash := sha512.Sum384(payload)

		r := new(big.Int)
		r.SetBytes(signature[0:48])

		s := new(big.Int)
		s.SetBytes(signature[48:96])

		return ecdsa.Verify(publicKeyInst.(*ecdsa.PublicKey), payloadHash[:], r, s), nil
	case RSA2048RESTR:
		return false, errors.New("RSA2048RESTR is not currently implemented!")
	case RSAPKCS:
		return false, errors.New("RSAPKCS is not currently implemented!")
	case RSAPSS:
		return false, errors.New("RSAPSS is not currently implemented!")
	default:
		return false, fmt.Errorf("PublicKey type %d is not supported!", pkType)
	}
}

func VerifyCoseSignature(coseSig CoseSignature, publicKey FdoPublicKey) (bool, error) {
	coseSigPayloadBytes, err := NewSig1Payload(coseSig.Protected, coseSig.Payload)
	if err != nil {
		return false, err
	}

	switch publicKey.PkEnc {
	case Crypto:
		return false, errors.New("EPID signatures are not currently supported!")
	case X509:
		pubKeyInst, err := x509.ParsePKIXPublicKey(publicKey.PkBody.([]byte))
		if err != nil {
			return false, errors.New("Error parsing PKIX X509 Public Key. " + err.Error())
		}

		signatureIsValid, err := VerifySignature(coseSigPayloadBytes, coseSig.Signature, pubKeyInst, publicKey.PkType)
		if err != nil {
			return false, err
		}

		return signatureIsValid, nil
	case X5CHAIN:
		var decCertBytes [][]byte

		for _, certBytesInst := range publicKey.PkBody.([]interface{}) {
			decCertBytes = append(decCertBytes, certBytesInst.([]byte))
		}

		chainIsValid, successChain, err := VerifyCertificateChain(decCertBytes)
		if err != nil {
			return false, err
		}

		if !chainIsValid {
			return false, errors.New("Failed to verify pkBody x5chain!")
		}

		leafCert := successChain[0]

		signatureIsValid, err := VerifySignature(coseSigPayloadBytes, coseSig.Signature, leafCert.PublicKey, publicKey.PkType)
		if err != nil {
			return false, err
		}

		return signatureIsValid, nil
	case COSEKEY:
		return false, errors.New("CoseKey is not currently supported!") // TODO
	default:
		return false, fmt.Errorf("PublicKey encoding %d is not supported!", publicKey.PkEnc)
	}
}

func GetDeviceSgType(pkType FdoPkType, hashType HashType) (DeviceSgType, error) {
	switch pkType {
	case SECP256R1:
		return StSECP256R1, nil
	case SECP384R1:
		return StSECP384R1, nil
	case RSA2048RESTR, RSAPKCS, RSAPSS:
		if hashType == FDO_SHA256 {
			return StRSA2048, nil
		} else if hashType == FDO_SHA384 {
			return StRSA3072, nil
		} else {
			return 0, fmt.Errorf("For RSA: %d is an unsupported hash type!", hashType)
		}
	default:
		return 0, fmt.Errorf("For RSA: %d is an unsupported public key type!", pkType)
	}
}

func ExtractPrivateKey(privateKeyDer []byte) (interface{}, error) {
	if key, err := x509.ParsePKCS1PrivateKey(privateKeyDer); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(privateKeyDer); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("Found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(privateKeyDer); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("Failed to parse private key")
}

func GenerateCoseSignature(payload []byte, protected ProtectedHeader, unprotected UnprotectedHeader, privateKeyInterface interface{}, sgType DeviceSgType) (*CoseSignature, error) {
	protectedBytes, err := cbor.Marshal(protected)
	coseSigPayloadBytes, err := NewSig1Payload(protectedBytes, payload)
	if err != nil {
		return nil, err
	}

	var signature []byte

	switch sgType {
	case StSECP256R1:
		payloadHash := sha256.Sum256(coseSigPayloadBytes)
		r, s, err := ecdsa.Sign(rand.Reader, privateKeyInterface.(*ecdsa.PrivateKey), payloadHash[:])
		if err != nil {
			return nil, errors.New("Error generating ES256 cose signature. " + err.Error())
		}

		signature = append(r.Bytes(), s.Bytes()...)
	case StSECP384R1:
		payloadHash := sha512.Sum384(coseSigPayloadBytes)
		r, s, err := ecdsa.Sign(rand.Reader, privateKeyInterface.(*ecdsa.PrivateKey), payloadHash[:])
		if err != nil {
			return nil, errors.New("Error generating ES384 cose signature. " + err.Error())
		}

		signature = append(r.Bytes(), s.Bytes()...)
	case StRSA3072:
		return nil, errors.New("StRSA3072 is not currently implemented!")
	case StRSA2048:
		return nil, errors.New("StRSA2048 is not currently implemented!")
	case StEPID10:
		return nil, errors.New("StEPID10 is not currently implemented!")
	default:
		return nil, fmt.Errorf("Alg %d is not supported!", sgType)
	}

	return &CoseSignature{
		Protected:   protectedBytes,
		Unprotected: unprotected,
		Payload:     payload,
		Signature:   signature,
	}, nil
}
