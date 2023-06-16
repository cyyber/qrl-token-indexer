package models

import "github.com/cyyber/qrl-token-indexer/common"

type TokenHolder struct {
	TokenTxHash common.Hash    `json:"tokenTxHash" bson:"tokenTxHash"`
	Address     common.Address `json:"address" bson:"address"`
	Amount      int64          `json:"amount" bson:"amount"`
}

func NewTokenHolder(tokenTxHash common.Hash, address common.Address, amount int64) *TokenHolder {
	return &TokenHolder{
		TokenTxHash: tokenTxHash,
		Address:     address,
		Amount:      amount,
	}
}
