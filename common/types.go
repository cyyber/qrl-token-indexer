package common

import "encoding/hex"

type Address [39]byte
type Hash [32]byte

func (a *Address) ToString() string {
	return hex.EncodeToString(a[:])
}

func (h *Hash) ToString() string {
	return hex.EncodeToString(h[:])
}
