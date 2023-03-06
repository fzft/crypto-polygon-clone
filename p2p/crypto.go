package p2p

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io"
)

func aesEncrypt(key []byte, plaintext []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext
}

func aesDecrypt(key []byte, ciphertextBytes []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertextBytes) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertextBytes, ciphertextBytes)
	return ciphertextBytes
}

func newkey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return privateKey
}

func rsaEncrypt(message []byte, pubKey []byte) []byte {
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKey,
	}
	pemEncoded := pem.EncodeToMemory(pemBlock)
	pemBlock, _ = pem.Decode(pemEncoded)
	// Parse the ASN.1 encoded public key
	publicKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err !=nil {
		panic(err)
	}
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey.(*rsa.PublicKey), message, nil)
	if err != nil {
		panic(err)
	}
	return ciphertext
}

func rsaDecrypt(message []byte, privateKey *rsa.PrivateKey) []byte {
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, message, nil)
	if err != nil {
		panic(err)
	}
	return plaintext
}
