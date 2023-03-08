package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	prvKey := GeneratePrivateKey()
	pubKey := prvKey.PublicKey()
	address := pubKey.Address()

	t.Log(address)
}

func TestKeypairSignVerify(t *testing.T) {
	prvKey := GeneratePrivateKey()
	pubKey := prvKey.PublicKey()
	//address := pubKey.Address()

	msg := []byte("hello world")
	sig, err := prvKey.Sign(msg)
	assert.Nil(t, err)
	assert.True(t, sig.Verify(pubKey, msg))
}
