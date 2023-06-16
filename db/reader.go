package db

import (
	"errors"
	"github.com/cyyber/qrl-token-indexer/common"
	"github.com/cyyber/qrl-token-indexer/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDBProcessor) GetLastBlock() (*models.Block, error) {
	o := &options.FindOneOptions{}
	o.Sort = bson.D{{"number", -1}}
	result := m.blocksCollection.FindOne(m.ctx, bson.D{{}}, o)
	if result.Err() != nil {
		return nil, result.Err()
	}
	b := &models.Block{}
	err := result.Decode(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (m *MongoDBProcessor) GetBlockByNumber(number int64) (*models.Block, error) {
	o := &options.FindOneOptions{}
	o.Sort = bson.D{{"number", -1}}
	result := m.blocksCollection.FindOne(m.ctx,
		bson.D{{"number", number}}, o)
	if result.Err() != nil {
		return nil, result.Err()
	}
	b := &models.Block{}
	err := result.Decode(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (m *MongoDBProcessor) GetTokenTxsByBlockNumber(blockNumber int64) ([]*models.TokenTx, error) {
	var tokenTxs []*models.TokenTx

	o := &options.FindOptions{}
	o.Sort = bson.D{{"blockNumber", -1}}
	cursor, err := m.tokenTxsCollection.Find(m.ctx,
		bson.D{{"blockNumber", blockNumber}}, o)
	if err != nil {
		return nil, err
	}
	for cursor.Next(m.ctx) {
		t := &models.TokenTx{}
		err := cursor.Decode(t)
		if err != nil {
			return nil, err
		}
		tokenTxs = append(tokenTxs, t)
	}

	return tokenTxs, nil
}

func (m *MongoDBProcessor) GetTransferTokenTxsByBlockNumber(blockNumber int64) ([]*models.TransferTokenTx, error) {
	var transferTokenTxs []*models.TransferTokenTx

	o := &options.FindOptions{}
	o.Sort = bson.D{{"blockNumber", -1}}
	cursor, err := m.transferTokenTxsCollection.Find(m.ctx,
		bson.D{{"blockNumber", blockNumber}}, o)
	if err != nil {
		return nil, err
	}
	for cursor.Next(m.ctx) {
		t := &models.TransferTokenTx{}
		err := cursor.Decode(t)
		if err != nil {
			return nil, err
		}
		transferTokenTxs = append(transferTokenTxs, t)
	}

	return transferTokenTxs, nil
}

func (m *MongoDBProcessor) GetTokenHolder(tokenTxHash common.Hash, address common.Address) (*models.TokenHolder, error) {
	o := &options.FindOneOptions{}
	o.Sort = bson.D{{"tokenTxHash", -1}, {"address", -1}}
	result := m.tokenHoldersCollection.FindOne(m.ctx,
		bson.D{
			{"tokenTxHash", tokenTxHash},
			{"address", address},
		}, o)
	if result.Err() != nil {
		return nil, result.Err()
	}
	t := &models.TokenHolder{}
	err := result.Decode(t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (m *MongoDBProcessor) GetTokenHolders(transferTokenTx *models.TransferTokenTx) (models.TokenHolders, error) {
	tokenHolders := make(models.TokenHolders)

	fromTokenHolder, err := m.GetTokenHolder(transferTokenTx.TokenTxHash, transferTokenTx.From)
	if err != nil {
		return nil, err
	}
	tokenHolders[transferTokenTx.From] = fromTokenHolder

	for _, address := range transferTokenTx.Addresses {
		tokenHolder, err := m.GetTokenHolder(transferTokenTx.TokenTxHash, address)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, err
			}
			tokenHolder = models.NewTokenHolder(transferTokenTx.TokenTxHash, address, 0)
		}
		tokenHolders[address] = tokenHolder
	}
	return tokenHolders, nil
}

func (m *MongoDBProcessor) GetTokenHoldersWithCache(transferTokenTx *models.TransferTokenTx, cache models.TokenHoldersCache) (models.TokenHolders, error) {
	if cache == nil {
		return nil, errors.New("TokenHoldersCache required")
	}
	var err error

	tokenHolders := make(models.TokenHolders)

	fromTokenHolder := cache.Get(transferTokenTx.TokenTxHash, transferTokenTx.From)
	if fromTokenHolder == nil {
		fromTokenHolder, err = m.GetTokenHolder(transferTokenTx.TokenTxHash, transferTokenTx.From)
		if err != nil {
			return nil, err
		}
	}

	tokenHolders[transferTokenTx.From] = fromTokenHolder

	for _, address := range transferTokenTx.Addresses {
		tokenHolder := cache.Get(transferTokenTx.TokenTxHash, address)
		if tokenHolder == nil {
			tokenHolder, err = m.GetTokenHolder(transferTokenTx.TokenTxHash, address)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					return nil, err
				}
				tokenHolder = models.NewTokenHolder(transferTokenTx.TokenTxHash, address, 0)
			}
		}

		tokenHolders[address] = tokenHolder
	}
	return tokenHolders, nil
}
