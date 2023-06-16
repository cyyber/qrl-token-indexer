package models

import (
	"github.com/cyyber/qrl-token-indexer/common"
	"github.com/cyyber/qrl-token-indexer/generated"
	"github.com/cyyber/qrl-token-indexer/misc"
	"github.com/cyyber/qrl-token-indexer/xmss"
)

type TransferTokenTx struct {
	BlockNumber   int64                    `json:"blockNumber" bson:"blockNumber"`
	TxHash        common.Hash              `json:"txHash" bson:"txHash"`
	TokenTxHash   common.Hash              `json:"tokenTxHash" bson:"tokenTxHash"`
	From          common.Address           `json:"from" bson:"from"`
	AddressAmount map[common.Address]int64 `json:"addressAmount" bson:"addressAmount"`
}

func (t *TransferTokenTx) GetTokenRelatedTx() *TokenRelatedTx {
	return NewTokenRelatedTx(t.TokenTxHash, t.TxHash)
}

func NewTransferTokenTxFromPBData(blockNumber uint64, pbData *generated.Transaction) *TransferTokenTx {
	tt := pbData.GetTransferToken()

	t := &TransferTokenTx{}
	t.BlockNumber = int64(blockNumber)
	t.TokenTxHash = misc.ToSizedHash(tt.TokenTxhash)
	t.TxHash = misc.ToSizedHash(pbData.TransactionHash)
	if pbData.MasterAddr != nil {
		t.From = xmss.GetXMSSAddressFromPK(pbData.MasterAddr)
	} else {
		t.From = xmss.GetXMSSAddressFromPK(pbData.PublicKey)
	}

	for i, addrTo := range tt.AddrsTo {
		sizedAddrTo := misc.ToSizedAddress(addrTo)
		t.AddressAmount[sizedAddrTo] += int64(tt.Amounts[i])
	}

	return t
}
