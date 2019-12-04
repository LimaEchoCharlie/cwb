package noise

import (
	"encoding/binary"
	"fmt"
	"github.com/flynn/noise"
)

var (
	diffieHellman     = noise.DH25519
	cipher            = noise.CipherAESGCM
	hash              = noise.HashSHA512
	sessionIDEncoding = binary.BigEndian
)

// EncryptionSessionID is assigned by the server to each client and returned to the client at the handshake completion.
// The client must pass it back to the server with every subsequent message that uses the cipher states from the handshake.
type EncryptionSessionID uint16

const (
	HandshakeSessionID = EncryptionSessionID(0)
	encryptionSessionIDSize = 2

)

// bytes encodes the EncryptionSessionID into a byte array
func (id EncryptionSessionID) bytes() []byte {
	b := make([]byte, encryptionSessionIDSize)
	sessionIDEncoding.PutUint16(b, uint16(id))
	return b
}

// convertByteSequence converts a byte array into an EncryptionSessionID
func convertByteSequence(b []byte) (id EncryptionSessionID, ok bool) {
	if len(b) < encryptionSessionIDSize {
		return 0, false
	}
	return EncryptionSessionID(sessionIDEncoding.Uint16(b)), true
}

type CipherStatePair struct {
	Decrypter *noise.CipherState
	Encrypter *noise.CipherState
}

type ClientMessenger interface {
	Exchange(message []byte) (reply []byte, err error)
}

func ClientHandshake(client ClientMessenger) (id EncryptionSessionID, csPair CipherStatePair, err error) {

	cs := noise.NewCipherSuite(diffieHellman, cipher, hash)

	handshakeState, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite: cs,
		Pattern:     noise.HandshakeNN,
		Initiator:   true,
	})
	msg, _, _, err := handshakeState.WriteMessage(nil, nil)
	if err != nil {
		return
	}
	encryptedReply, err := client.Exchange(msg)
	if err != nil {
		return
	}
	var serverResponse []byte
	serverResponse, csPair.Encrypter, csPair.Decrypter, err = handshakeState.ReadMessage(nil, encryptedReply)
	if err != nil {
		return
	}
	var ok bool
	if id, ok = convertByteSequence(serverResponse); !ok {
		err = fmt.Errorf("unable to deduce encryption session id")
	}
	return
}

type ServerMessenger interface {
	Send(message []byte) (err error)
}

func ServerHandshake(server ServerMessenger, id EncryptionSessionID, initiator []byte) (csPair CipherStatePair, err error) {

	cs := noise.NewCipherSuite(diffieHellman, cipher, hash)

	handshakeState, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite: cs,
		Pattern:     noise.HandshakeNN,
		Initiator:   false,
	})
	_, _, _, err = handshakeState.ReadMessage(nil, initiator)
	if err != nil {
		return csPair, err
	}

	var encodedReply []byte
	encodedReply, csPair.Decrypter, csPair.Encrypter, err = handshakeState.WriteMessage(nil, id.bytes())
	if err != nil {
		return csPair, err
	}
	err = server.Send(encodedReply)
	if err != nil {
		return csPair, err
	}
	return csPair, nil
}
