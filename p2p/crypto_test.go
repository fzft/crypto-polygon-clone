package p2p

import (
	"crypto/md5"
	"crypto/x509"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAes(t *testing.T) {

	privateKey := newkey()
	pubKey := privateKey.PublicKey
	pubKeyBytes, _  := x509.MarshalPKIXPublicKey(&pubKey)
	out := md5.Sum([]byte(pubKeyBytes))
	t.Log(len(out))
}

func TestRsa(t *testing.T) {
	privateKey := newkey()
	pubKey := privateKey.PublicKey
	message := []byte("hello")
	pubKeyBytes, _  := x509.MarshalPKIXPublicKey(&pubKey)

	t.Log(len(pubKeyBytes))
	cipherText := rsaEncrypt(message, pubKeyBytes)
	plainText := rsaDecrypt(cipherText, privateKey)

	assert.Equal(t, message, plainText)
}