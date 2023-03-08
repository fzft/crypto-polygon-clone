package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/fzft/crypto-simple-blockchain/types"
	"math/big"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

func GeneratePrivateKey() PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return PrivateKey{
		key: key,
	}
}

func (k PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, k.key, data)
	if err !=nil {
		panic(err)
	}

	return &Signature{r, s}, nil
}

func (k PrivateKey) PublicKey() PublicKey {
	return PublicKey{key: &k.key.PublicKey}
}

type PublicKey struct {
	key *ecdsa.PublicKey
}

func (k PublicKey) ToSlice() []byte {
	return elliptic.Marshal(k.key.Curve, k.key.X, k.key.Y)
}

func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k.ToSlice())
	return types.NewAddressFromBytes(h[len(h)-20:])
}


type Signature struct {
	r, s *big.Int
}

func (sig Signature) Verify(pubKey PublicKey, data []byte) bool {
	return ecdsa.Verify(pubKey.key, data, sig.r, sig.s)
}
