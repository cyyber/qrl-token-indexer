package models

import (
	"fmt"

	"github.com/cyyber/qrl-token-indexer/common"
)

type TokenHoldersCache map[string]*TokenHolder

func (t TokenHoldersCache) Get(tokenTxHash common.Hash, address common.Address) *TokenHolder {
	key := t.getKey(tokenTxHash, address)

	return t[key]
}

func (t TokenHoldersCache) Put(tokenTxHash common.Hash, address common.Address, value *TokenHolder) {
	key := t.getKey(tokenTxHash, address)

	t[key] = value
}

func (t TokenHoldersCache) Union(tokenTxHash common.Hash, address common.Address, value *TokenHolder) {
	key := t.getKey(tokenTxHash, address)
	if _, ok := t[key]; ok {
		return
	}
	t[key] = value
}

func (t TokenHoldersCache) PutFromTokenHolders(tokenHolders TokenHolders) {
	for _, tokenHolder := range tokenHolders {
		t.Put(tokenHolder.TokenTxHash, tokenHolder.Address, tokenHolder)
	}
}

func (t TokenHoldersCache) UnionFromTokenHolders(tokenHolders TokenHolders) {
	for _, tokenHolder := range tokenHolders {
		t.Put(tokenHolder.TokenTxHash, tokenHolder.Address, tokenHolder)
	}
}

func (t TokenHoldersCache) getKey(tokenTxHash common.Hash, address common.Address) string {
	return fmt.Sprintf("%s_%s",
		tokenTxHash.ToString(), address.ToString())
}
