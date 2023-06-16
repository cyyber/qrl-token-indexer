package models

import (
	"github.com/cyyber/qrl-token-indexer/common"
	"github.com/cyyber/qrl-token-indexer/generated"
	"github.com/cyyber/qrl-token-indexer/misc"
)

type TokenTx struct {
	BlockNumber     int64                    `json:"blockNumber" bson:"blockNumber"`
	TxHash          common.Hash              `json:"txHash" bson:"txHash"`
	Name            []byte                   `json:"name" bson:"name"`
	Decimals        int64                    `json:"decimals" bson:"decimals"`
	InitialBalances map[common.Address]int64 `json:"initialBalances" bson:"initialBalances"`
}

func (t *TokenTx) GetTokenHolders() TokenHolders {
	tokenHolders := make(TokenHolders)
	for address, balance := range t.InitialBalances {
		tokenHolder := NewTokenHolder(t.TxHash, address, balance)
		tokenHolders[address] = tokenHolder
	}
	return tokenHolders
}

func NewTokenTxFromPBData(blockNumber uint64, pbData *generated.Transaction) *TokenTx {
	tt := pbData.GetToken()

	t := &TokenTx{}
	t.BlockNumber = int64(blockNumber)
	t.TxHash = misc.ToSizedHash(pbData.TransactionHash)
	t.Name = tt.Name
	t.Decimals = int64(tt.Decimals)
	t.InitialBalances = make(map[common.Address]int64)

	for _, addressAmount := range tt.InitialBalances {
		sizedAddrTo := misc.ToSizedAddress(addressAmount.Address)
		t.InitialBalances[sizedAddrTo] += int64(addressAmount.Amount)
	}

	return t
}
