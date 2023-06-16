package models

import "github.com/cyyber/qrl-token-indexer/common"

type TokenRelatedTx struct {
	TokenTxHash common.Hash `json:"tokenTxHash" bson:"tokenTxHash"`
	TxHash      common.Hash `json:"txHash" bson:"txHash"`
}

func NewTokenRelatedTx(tokenTxHash, txHash common.Hash) *TokenRelatedTx {
	return &TokenRelatedTx{
		TokenTxHash: tokenTxHash,
		TxHash:      txHash,
	}
}
