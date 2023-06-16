package models

import (
	"encoding/hex"
	"fmt"
	"github.com/cyyber/qrl-token-indexer/common"
)

type TokenHolders map[common.Address]*TokenHolder

func (t TokenHolders) Apply(tx *TransferTokenTx) error {
	from := tx.From
	fromTokenHolder, ok := t[from]
	if !ok {
		return fmt.Errorf("from address: %s not found for TokenHolder", from.ToString())
	}

	for address, amount := range tx.AddressAmount {
		if uint64(fromTokenHolder.Amount) < uint64(amount) {
			return fmt.Errorf("from address: %s "+
				"txhash: %s "+
				"doesn't have sufficient token: %s "+
				"remaining %d but trying to spend %d",
				from.ToString(),
				tx.TxHash.ToString(),
				tx.TokenTxHash.ToString(),
				uint64(fromTokenHolder.Amount), uint64(amount))
		}
		fromTokenHolder.Amount -= amount
		tokenHolder, ok := t[address]
		if !ok {
			return fmt.Errorf("address: %s "+
				"txhash: %s "+
				"token: %s is not found in tokenHolders map",
				from.ToString(),
				tx.TxHash.ToString(),
				tx.TokenTxHash.ToString())
		}
		tokenHolder.Amount += amount
	}

	return nil
}

func (t TokenHolders) Revert(tx *TransferTokenTx) error {
	from := tx.From
	fromTokenHolder, ok := t[from]
	if !ok {
		return fmt.Errorf("from address: %s not found for TokenHolder",
			hex.EncodeToString(from[:]))
	}

	for address, amount := range tx.AddressAmount {
		tokenHolder, ok := t[address]
		if !ok {
			return fmt.Errorf("address: %s "+
				"txhash: %s "+
				"token: %s is not found in tokenHolders map",
				from.ToString(),
				tx.TxHash.ToString(),
				tx.TokenTxHash.ToString())
		}

		if uint64(tokenHolder.Amount) < uint64(amount) {
			return fmt.Errorf("from address: %s "+
				"txhash: %s "+
				"doesn't have sufficient token: %s "+
				"remaining %d but trying to spend %d",
				from.ToString(),
				tx.TxHash.ToString(),
				tx.TokenTxHash.ToString(),
				uint64(tokenHolder.Amount), uint64(amount))
		}

		tokenHolder.Amount -= amount
		fromTokenHolder.Amount += amount
	}

	return nil
}
