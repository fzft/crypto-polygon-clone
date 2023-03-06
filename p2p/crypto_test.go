package p2p

import (
	"crypto/x509"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAes(t *testing.T) {
}

func TestRsa(t *testing.T) {
	privateKey := newkey()
	pubKey := privateKey.PublicKey
	message := []byte("hello")
	pubKeyBytes, _  := x509.MarshalPKIXPublicKey(&pubKey)
	cipherText := rsaEncrypt(message, pubKeyBytes)
	plainText := rsaDecrypt(cipherText, privateKey)

	assert.Equal(t, message, plainText)
}