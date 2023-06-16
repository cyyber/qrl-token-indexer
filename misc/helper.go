package misc

import (
	"crypto/sha256"

	"github.com/cyyber/qrl-token-indexer/common"
	"golang.org/x/crypto/sha3"
)

func ToSizedHash(hash []byte) common.Hash {
	var sizedHash common.Hash
	copy(sizedHash[:], hash)
	return sizedHash
}

func ToSizedAddress(address []byte) common.Address {
	var sizedAddress common.Address
	copy(sizedAddress[:], address)
	return sizedAddress
}

func SHAKE128(out, msg []byte) []byte {
	hasher := sha3.NewShake128()
	hasher.Write(msg)
	hasher.Read(out)
	return out
}

func SHAKE256(out, msg []byte) []byte {
	hasher := sha3.NewShake256()
	hasher.Write(msg)
	hasher.Read(out)
	return out
}

func SHA256(out, msg []byte) []byte {
	hasher := sha256.New()
	hasher.Write(msg)
	hashOut := hasher.Sum(nil)
	copy(out, hashOut)
	return out
}
