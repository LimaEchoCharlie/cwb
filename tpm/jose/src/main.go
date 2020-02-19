package main

import (
	"crypto"
	"encoding/json"
	"fmt"
	"github.com/google/go-tpm-tools/simulator"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
	"gopkg.in/square/go-jose.v2"
	"io"
	"log"
	"time"
)

const keyPassword = ""

// publicKey reads the public key from the TPM
func publicKey(rw io.ReadWriter, handle tpmutil.Handle) (crypto.PublicKey, error) {
	public, _, _, err := tpm2.ReadPublic(rw, handle)
	if err != nil {
		return nil, err
	}
	return public.Key()
}

// joseSigner is a TPM based signer that implements the JOSE opaque signer interface
type joseSigner struct {
	rw     io.ReadWriter
	handle tpmutil.Handle
}

// Public returns the public key of the current signing key.
func (s joseSigner) Public() *jose.JSONWebKey {
	key, err := publicKey(s.rw, s.handle)
	if err != nil {
		return nil
	}
	return &jose.JSONWebKey{Key: key, Algorithm: string(jose.ES256), Use: "sig"}
}

// Algs returns a list of supported signing algorithms.
func (s joseSigner) Algs() []jose.SignatureAlgorithm {
	return []jose.SignatureAlgorithm{jose.ES256}
}

// SignPayload hashes and signs a payload
func (s joseSigner) SignPayload(payload []byte, alg jose.SignatureAlgorithm) ([]byte, error) {
	if alg != jose.ES256 {
		return nil, fmt.Errorf("unsupported algorithm %v", alg)
	}
	digest, err := tpm2.Hash(s.rw, tpm2.AlgSHA256, payload)
	if err != nil {
		return nil, err
	}

	sig, err := tpm2.Sign(s.rw, s.handle, keyPassword, digest, nil)
	if err != nil {
		return nil, err
	}
	//signature, err := asn1.Marshal(
	//	struct {
	//		R, S *big.Int
	//	}{sig.ECC.R, sig.ECC.S},
	//)

	// taken from ecDecrypterSigner.signPayload in gopkg.in/square/go-jose.v2
	// if asn1 marshalling is used, then the verification will fail
	keyBytes := 32
	rBytes := sig.ECC.R.Bytes()
	rBytesPadded := make([]byte, keyBytes)
	copy(rBytesPadded[keyBytes-len(rBytes):], rBytes)

	sBytes := sig.ECC.S.Bytes()
	sBytesPadded := make([]byte, keyBytes)
	copy(sBytesPadded[keyBytes-len(sBytes):], sBytes)

	out := append(rBytesPadded, sBytesPadded...)
	return out, err
}

// initTPM initialises the TPM simulator and creates a signing key
func initTPM() (rwc io.ReadWriteCloser, key tpmutil.Handle, err error) {
	rwc, err = simulator.GetWithFixedSeedInsecure(0)
	if err != nil {
		return rwc, key, err
	}
	defer rwc.Close()

	signingTemplate := tpm2.Public{
		Type:       tpm2.AlgECC,
		NameAlg:    tpm2.AlgSHA256,
		Attributes: tpm2.FlagSign | tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth,
		ECCParameters: &tpm2.ECCParams{
			Sign: &tpm2.SigScheme{
				Alg:  tpm2.AlgECDSA,
				Hash: tpm2.AlgSHA256,
			},
			CurveID: tpm2.CurveNISTP256,
		}}
	key, _, err = tpm2.CreatePrimary(rwc, tpm2.HandleOwner, tpm2.PCRSelection{}, "", keyPassword, signingTemplate)
	if err != nil {
		return rwc, key, err
	}
	return rwc, key, err
}

// signJWT creates a signed JWT using the TPM signing key
func signJWT(rw io.ReadWriter, keyHandle tpmutil.Handle, payload interface{}) (jwt string, err error) {

	opaqueSigner := joseSigner{rw: rw, handle: keyHandle}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: opaqueSigner}, nil)
	if err != nil {
		return jwt, err
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return jwt, err
	}
	object, err := signer.Sign(b)
	if err != nil {
		return jwt, err
	}
	return object.CompactSerialize()
}

// verifyJWT parses and verifies the JWT using the supplied public key
func verifyJWT(public crypto.PublicKey, jwt string) (p []byte, err error) {
	signature, err := jose.ParseSigned(jwt)
	if err != nil {
		return p, err
	}
	p, err = signature.Verify(public)
	if err != nil {
		return p, err
	}
	return p, err
}

type token struct {
	Exp int64 `json:"exp"`
}

func (t token) String() string {
	return fmt.Sprintf("{ exp : %d }", t.Exp)
}

// tpmJWTExample shows how to sign a JWT using a key in a TPM
func tpmJWTExample() (err error) {

	fmt.Println("Initialise TPM")
	rwc, keyHandle, err := initTPM()
	if err != nil {
		return err
	}
	defer func() {
		tpm2.FlushContext(rwc, keyHandle)
		rwc.Close()
	}()

	fmt.Println("Read public key from TPM")
	public, err := publicKey(rwc, keyHandle)
	if err != nil {
		return err
	}

	payload := token{time.Now().Add(100 * 365 * 24 * time.Hour).Unix()}
	fmt.Printf("Example payload: %v\n", payload)

	fmt.Println("Create and sign JWT")
	jwt, err := signJWT(rwc, keyHandle, payload)
	if err != nil {
		return err
	}
	fmt.Printf("JWT: %v\n", jwt)

	fmt.Println("Verifying JWT")
	// parse and verify JWT
	receivedBytes, err := verifyJWT(public, jwt)
	if err != nil {
		return err
	}
	fmt.Println("Verified")

	fmt.Println("Checking that the payload is intact")
	var receivedPayload token
	if err = json.Unmarshal(receivedBytes, &receivedPayload); err != nil {
		return err
	}
	if payload != receivedPayload {
		return fmt.Errorf("payloads mismatch %v:%v", payload, receivedPayload)
	}

	return err
}
func main() {
	err := tpmJWTExample()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("SUCCESS")
}
