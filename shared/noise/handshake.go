package noise

import (
	"github.com/flynn/noise"
)

type ClientMessenger interface {
	SendReceive(message []byte) (reply []byte, err error)
}

func ClientHandshake(client ClientMessenger) (encrypter *noise.CipherState, decrypter *noise.CipherState, err error) {

	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherAESGCM, noise.HashSHA512)

	handshakeState, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite: cs,
		Pattern:     noise.HandshakeNN,
		Initiator:   true,
	})
	msg, _, _, err := handshakeState.WriteMessage(nil, nil)
	if err != nil {
		return nil, nil, err
	}
	encryptedReply, err := client.SendReceive(msg)
	if err != nil {
		return nil, nil, err
	}
	_, encrypter, decrypter, err = handshakeState.ReadMessage(nil, encryptedReply)
	if err != nil {
		return encrypter, decrypter, err
	}
	return encrypter, decrypter, err
}

type ServerMessenger interface {
	Receive() (message []byte, err error)
	Send(message []byte) (err error)
}

func ServerHandshake(server ServerMessenger) (decrypter *noise.CipherState, encrypter *noise.CipherState, err error) {

	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherAESGCM, noise.HashSHA512)

	handshakeState, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite: cs,
		Pattern:     noise.HandshakeNN,
		Initiator:   false,
	})
	encodedMessage, err := server.Receive()
	if err != nil {
		return nil, nil, err
	}
	_, _, _, err = handshakeState.ReadMessage(nil, encodedMessage)
	if err != nil {
		return nil, nil, err
	}

	encodedReply, decrypter, encrypter, err := handshakeState.WriteMessage(nil, nil)
	if err != nil {
		return nil, nil, err
	}
	err = server.Send(encodedReply)
	if err != nil {
		return nil, nil, err
	}
	return decrypter, encrypter, nil
}
