package noise

import (
	"encoding/binary"
	"github.com/flynn/noise"
)

var (
	diffieHellman     = noise.DH25519
	cipher            = noise.CipherAESGCM
	hash              = noise.HashSHA512
	sessionIDEncoding = binary.BigEndian
)

// ChannelID uniquely identifies the channel between a client and a server
// The client must pass it back to the server with every subsequent message that uses the cipher states from the handshake.
type ChannelID []byte

const (
	hashByteLen = 8
)

// UInt64 converts a ChannelID into an uint64
func (id ChannelID) UInt64() (uint64, bool) {
	if len(id) < hashByteLen {
		return 0, false
	}
	return sessionIDEncoding.Uint64(id), true
}

type CipherStatePair struct {
	Decrypter *noise.CipherState
	Encrypter *noise.CipherState
}

type ClientMessenger interface {
	Exchange(message []byte) (reply []byte, err error)
}

func ClientHandshake(client ClientMessenger) (id ChannelID, csPair CipherStatePair, err error) {

	cs := noise.NewCipherSuite(diffieHellman, cipher, hash)

	handshakeState, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite: cs,
		Pattern:     noise.HandshakeNN,
		Initiator:   true,
	})
	msg, _, _, err := handshakeState.WriteMessage(nil, nil)
	if err != nil {
		return id, csPair, err
	}
	encryptedReply, err := client.Exchange(msg)
	if err != nil {
		return id, csPair, err
	}
	_, csPair.Encrypter, csPair.Decrypter, err = handshakeState.ReadMessage(nil, encryptedReply)
	if err != nil {
		return id, csPair, err
	}
	return handshakeState.ChannelBinding(), csPair, nil
}

type ServerMessenger interface {
	Send(message []byte) (err error)
}

func ServerHandshake(server ServerMessenger, initiator []byte) (id ChannelID, csPair CipherStatePair, err error) {

	cs := noise.NewCipherSuite(diffieHellman, cipher, hash)

	handshakeState, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite: cs,
		Pattern:     noise.HandshakeNN,
		Initiator:   false,
	})
	_, _, _, err = handshakeState.ReadMessage(nil, initiator)
	if err != nil {
		return id, csPair, err
	}

	var encodedReply []byte
	encodedReply, csPair.Decrypter, csPair.Encrypter, err = handshakeState.WriteMessage(nil, nil)
	if err != nil {
		return id, csPair, err
	}
	err = server.Send(encodedReply)
	if err != nil {
		return id, csPair, err
	}
	return handshakeState.ChannelBinding(), csPair, nil
}
