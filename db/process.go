package db

import (
	"encoding/hex"
	"github.com/cyyber/qrl-token-indexer/db/models"
	"github.com/cyyber/qrl-token-indexer/generated"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

func AddInsertOneModelIntoOperations(operations []mongo.WriteModel, model interface{}) {
	operation := mongo.NewInsertOneModel()
	operation.SetDocument(model)
	operations = append(operations, operation)
}

func AddDeleteOneModelIntoOperations(operations []mongo.WriteModel, model interface{}) {
	operation := mongo.NewDeleteOneModel()
	operation.SetFilter(model)
	operations = append(operations, operation)
}

func (m *MongoDBProcessor) ProcessBlock(b *generated.Block) error {
	var blockOperations []mongo.WriteModel
	var tokenTxOperations []mongo.WriteModel
	var transferTokenTxOperations []mongo.WriteModel
	var tokenHolderOperations []mongo.WriteModel
	var tokenRelatedTxOperations []mongo.WriteModel

	blockModel := models.NewBlockFromPBData(b)
	AddInsertOneModelIntoOperations(blockOperations, blockModel)

	tokenHoldersCache := make(models.TokenHoldersCache)

	for _, protoTX := range b.Transactions {
		switch protoTX.TransactionType.(type) {
		case *generated.Transaction_Token_:
			tokenTx := models.NewTokenTxFromPBData(b.Header.BlockNumber, protoTX)
			AddInsertOneModelIntoOperations(tokenTxOperations, tokenTx)

			tokenHolders := tokenTx.GetTokenHolders()
			tokenHoldersCache.PutFromTokenHolders(tokenHolders)

			AddInsertOneModelIntoOperations(tokenHolderOperations, tokenHolders)
		case *generated.Transaction_TransferToken_:
			transferTokenTx := models.NewTransferTokenTxFromPBData(b.Header.BlockNumber, protoTX)
			AddInsertOneModelIntoOperations(transferTokenTxOperations, transferTokenTxOperations)

			AddInsertOneModelIntoOperations(tokenRelatedTxOperations, transferTokenTx.GetTokenRelatedTx())

			tokenHolders, err := m.GetTokenHoldersWithCache(transferTokenTx, tokenHoldersCache)
			if err != nil {
				m.log.Error("[ProcessBlock] Error calling GetTokenHolders",
					"Error", err.Error())
				return err
			}
			err = tokenHolders.Apply(transferTokenTx)
			if err != nil {
				m.log.Error("[ProcessBlock] Failed to process block",
					"#", b.Header.BlockNumber,
					"Hash", hex.EncodeToString(b.Header.HashHeader))
				return err
			}
			tokenHoldersCache.PutFromTokenHolders(tokenHolders)

			var operation *mongo.UpdateOneModel
			for _, tokenHolder := range tokenHolders {
				operation = mongo.NewUpdateOneModel()
				operation.SetUpsert(true)
				operation.SetFilter(bsonx.Doc{
					{"tokenTxHash", bsonx.Binary(0, tokenHolder.TokenTxHash[:])},
					{"address", bsonx.Binary(0, tokenHolder.Address[:])},
				})
				operation.SetUpdate(tokenHolder)
				tokenHolderOperations = append(tokenHolderOperations, operation)
			}

		default:
			continue
		}
	}

	session, err := m.client.StartSession(options.Session())
	if err != nil {
		m.log.Error("[ProcessBlock] failed to start session")
		return err
	}
	defer session.EndSession(m.ctx)

	err = mongo.WithSession(m.ctx, session, func(sctx mongo.SessionContext) error {
		if err := sctx.StartTransaction(); err != nil {
			return err
		}

		if _, err := m.blocksCollection.BulkWrite(sctx, blockOperations); err != nil {
			return err
		}
		if _, err := m.tokenTxsCollection.BulkWrite(sctx, tokenTxOperations); err != nil {
			return err
		}
		if _, err := m.transferTokenTxsCollection.BulkWrite(sctx, tokenTxOperations); err != nil {
			return err
		}
		if _, err := m.tokenHoldersCollection.BulkWrite(sctx, tokenTxOperations); err != nil {
			return err
		}
		if _, err := m.tokenRelatedTxsCollection.BulkWrite(sctx, tokenRelatedTxOperations); err != nil {
			return err
		}

		return sctx.CommitTransaction(sctx)
	})
	if err != nil {
		m.log.Info("Failed to Process",
			"Block #", b.Header.BlockNumber,
			"HeaderHash", hex.EncodeToString(b.Header.HashHeader),
			"Error", err)
		return err
	}

	m.log.Info("Processed",
		"Block #", b.Header.BlockNumber,
		"HeaderHash", hex.EncodeToString(b.Header.HashHeader))
	return nil
}

func (m *MongoDBProcessor) RevertLastBlock() error {
	b, err := m.GetLastBlock()
	if err != nil {
		m.log.Error("[RevertLastBlock] failed to get last block",
			"error", err)
		return err
	}

	tokenTxs, err := m.GetTokenTxsByBlockNumber(b.Number)
	if err != nil {
		m.log.Error("[RevertLastBlock] failed to get token txs by block number",
			"block number", b.Number,
			"error", err)
		return err
	}
	transferTokenTxs, err := m.GetTransferTokenTxsByBlockNumber(b.Number)
	if err != nil {
		m.log.Error("[RevertLastBlock] failed to get transfer token txs by block number",
			"block number", b.Number,
			"error", err)
		return err
	}

	var blockOperations []mongo.WriteModel
	var tokenTxOperations []mongo.WriteModel
	var transferTokenTxOperations []mongo.WriteModel
	var tokenHolderOperations []mongo.WriteModel
	var tokenRelatedTxOperations []mongo.WriteModel

	tokenHoldersCache := make(models.TokenHoldersCache)

	for i := len(transferTokenTxs) - 1; i >= 0; i-- {
		tokenHolders, err := m.GetTokenHoldersWithCache(transferTokenTxs[i], tokenHoldersCache)
		if err != nil {
			m.log.Error("[RevertLastBlock] Error calling GetTokenHolders",
				"Error", err.Error())
			return err
		}

		transferTokenTx := transferTokenTxs[i]
		err = tokenHolders.Revert(transferTokenTx)
		if err != nil {
			m.log.Error("[RevertLastBlock] Failed to revert block",
				"#", b.Number,
				"Hash", b.Hash.ToString())
			return err
		}
		tokenHoldersCache.PutFromTokenHolders(tokenHolders)

		var operation *mongo.UpdateOneModel
		var deleteOperation *mongo.DeleteOneModel
		for _, tokenHolder := range tokenHolders {
			if tokenHolder.Amount == 0 {
				deleteOperation = mongo.NewDeleteOneModel()
				deleteOperation.SetFilter(bsonx.Doc{
					{"tokenTxHash", bsonx.Binary(0, tokenHolder.TokenTxHash[:])},
					{"address", bsonx.Binary(0, tokenHolder.Address[:])},
				})
				tokenHolderOperations = append(tokenHolderOperations, deleteOperation)
			} else {
				operation = mongo.NewUpdateOneModel()
				operation.SetUpsert(true)
				operation.SetFilter(bsonx.Doc{
					{"tokenTxHash", bsonx.Binary(0, tokenHolder.TokenTxHash[:])},
					{"address", bsonx.Binary(0, tokenHolder.Address[:])},
				})
				operation.SetUpdate(tokenHolder)
				tokenHolderOperations = append(tokenHolderOperations, operation)
			}
		}
		deleteOperation = mongo.NewDeleteOneModel()
		deleteOperation.SetFilter(bsonx.Doc{
			{"tokenTxHash", bsonx.Binary(0, transferTokenTx.TokenTxHash[:])},
			{"txHash", bsonx.Binary(0, transferTokenTx.TxHash[:])},
		})
		tokenRelatedTxOperations = append(tokenRelatedTxOperations, deleteOperation)

		deleteOperation = mongo.NewDeleteOneModel()
		deleteOperation.SetFilter(bsonx.Doc{
			{"txHash", bsonx.Binary(0, transferTokenTx.TokenTxHash[:])},
		})
		transferTokenTxOperations = append(transferTokenTxOperations, deleteOperation)
	}

	for i := len(tokenTxs) - 1; i >= 0; i-- {
		tokenTx := tokenTxs[i]
		tokenHolders := tokenTx.GetTokenHolders()

		var deleteOperation *mongo.DeleteOneModel
		for _, tokenHolder := range tokenHolders {
			deleteOperation = mongo.NewDeleteOneModel()
			deleteOperation.SetFilter(bsonx.Doc{
				{"tokenTxHash", bsonx.Binary(0, tokenHolder.TokenTxHash[:])},
				{"address", bsonx.Binary(0, tokenHolder.Address[:])},
			})
			tokenHolderOperations = append(tokenHolderOperations, deleteOperation)
		}

		deleteOperation = mongo.NewDeleteOneModel()
		deleteOperation.SetFilter(bsonx.Doc{
			{"txHash", bsonx.Binary(0, tokenTx.TxHash[:])},
		})
		tokenTxOperations = append(tokenTxOperations, deleteOperation)
	}

	AddDeleteOneModelIntoOperations(blockOperations, b)

	session, err := m.client.StartSession(options.Session())
	if err != nil {
		m.log.Error("[RevertLastBlock] failed to start session")
		return err
	}
	defer session.EndSession(m.ctx)

	err = mongo.WithSession(m.ctx, session, func(sctx mongo.SessionContext) error {
		if err := sctx.StartTransaction(); err != nil {
			return err
		}

		if _, err := m.blocksCollection.BulkWrite(sctx, blockOperations); err != nil {
			return err
		}
		if _, err := m.tokenTxsCollection.BulkWrite(sctx, tokenTxOperations); err != nil {
			return err
		}
		if _, err := m.transferTokenTxsCollection.BulkWrite(sctx, tokenTxOperations); err != nil {
			return err
		}
		if _, err := m.tokenHoldersCollection.BulkWrite(sctx, tokenTxOperations); err != nil {
			return err
		}
		if _, err := m.tokenRelatedTxsCollection.BulkWrite(sctx, tokenRelatedTxOperations); err != nil {
			return err
		}

		return sctx.CommitTransaction(sctx)
	})
	if err != nil {
		m.log.Info("Failed to Revert",
			"Block #", b.Number,
			"HeaderHash", b.Hash.ToString(),
			"Error", err)
		return err
	}

	m.log.Info("Reverted",
		"Block #", b.Number,
		"HeaderHash", b.Hash.ToString())
	return nil
}
