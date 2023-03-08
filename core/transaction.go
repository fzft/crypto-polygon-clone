package core

import (
	"fmt"
	"github.com/fzft/crypto-polygon-clone/crypto"
)

type Transaction struct {
	Data []byte

	PublicKey crypto.PublicKey
	Signature *crypto.Signature
}

func (tx *Transaction) Sign(prv crypto.PrivateKey) error {
	sig, err := prv.Sign(tx.Data)
	if err != nil {
		return err
	}

	tx.PublicKey = prv.PublicKey()
	tx.Signature = sig
	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("no signature")
	}

	if !tx.Signature.Verify(tx.PublicKey, tx.Data) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
