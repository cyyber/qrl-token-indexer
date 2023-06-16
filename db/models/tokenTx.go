package models

import (
	"github.com/cyyber/qrl-token-indexer/common"
	"github.com/cyyber/qrl-token-indexer/generated"
	"github.com/cyyber/qrl-token-indexer/misc"
)

type TokenTx struct {
	BlockNumber int64            `json:"blockNumber" bson:"blockNumber"`
	TxHash      common.Hash      `json:"txHash" bson:"txHash"`
	Name        []byte           `json:"name" bson:"name"`
	Decimals    int64            `json:"decimals" bson:"decimals"`
	Addresses   []common.Address `json:"addresses" bson:"addresses"`
	Amounts     []int64          `json:"amounts" bson:"amounts"`
}

func (t *TokenTx) GetTokenHolders() TokenHolders {
	tokenHolders := make(TokenHolders)
	for i, address := range t.Addresses {
		tokenHolder := NewTokenHolder(t.TxHash, address, t.Amounts[i])
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

	t.Addresses = make([]common.Address, 0, len(tt.InitialBalances))
	t.Amounts = make([]int64, 0, len(tt.InitialBalances))

	for _, addressAmount := range tt.InitialBalances {
		sizedAddrTo := misc.ToSizedAddress(addressAmount.Address)
		t.Addresses = append(t.Addresses, sizedAddrTo)
		t.Amounts = append(t.Amounts, int64(addressAmount.Amount))
	}

	return t
}
