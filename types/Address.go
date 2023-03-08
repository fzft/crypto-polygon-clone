package types

import "encoding/hex"

type Address [20]uint8

func NewAddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		panic("invalid length")
	}

	var addr Address
	copy(addr[:], b)
	return addr
}

func (a Address) String() string {
	return hex.EncodeToString(a[:])
}
