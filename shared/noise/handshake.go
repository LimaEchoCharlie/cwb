package noise

import (
	"github.com/flynn/noise"
)

var (
	diffieHellman = noise.DH25519
	cipher = noise.CipherAESGCM
	hash = noise.HashSHA512
)

type CipherStatePair struct {
	Decrypter *noise.CipherState
	Encrypter *noise.CipherState
}

type ClientMessenger interface {
	Exchange(message []byte) (reply []byte, err error)
}

func ClientHandshake(client ClientMessenger) (csPair CipherStatePair, err error) {

	cs := noise.NewCipherSuite(diffieHellman, cipher, hash)

	handshakeState, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite: cs,
		Pattern:     noise.HandshakeNN,
		Initiator:   true,
	})
	msg, _, _, err := handshakeState.WriteMessage(nil, nil)
	if err != nil {
		return csPair, err
	}
	encryptedReply, err := client.Exchange(msg)
	if err != nil {
		return csPair, err
	}
	_, csPair.Encrypter, csPair.Decrypter, err = handshakeState.ReadMessage(nil, encryptedReply)
	if err != nil {
		return csPair, err
	}
	return csPair, err
}

type ServerMessenger interface {
	Send(message []byte) (err error)
}

func ServerHandshake(server ServerMessenger, initiator []byte) (csPair CipherStatePair, err error) {

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
	encodedReply, csPair.Decrypter, csPair.Encrypter, err = handshakeState.WriteMessage(nil, nil)
	if err != nil {
		return csPair, err
	}
	err = server.Send(encodedReply)
	if err != nil {
		return csPair, err
	}
	return csPair, nil
}
