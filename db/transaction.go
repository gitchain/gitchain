package db

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/types"
)

func (db *T) GetTransactionBlock(hash []byte) (*block.Block, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("blocks"))
	if bucket == nil {
		return nil, nil
	}
	blockHash := bucket.Get(append([]byte("T"), hash...))
	if blockHash == nil {
		return nil, errors.New(fmt.Sprintf("referenced block %v not found", blockHash))
	}
	b := bucket.Get(blockHash)
	decodedBlock, err := block.Decode(b)
	if err != nil {
		return decodedBlock, err
	}
	return decodedBlock, nil
}

func (db *T) GetTransactionConfirmations(hash []byte) (int, error) {
	path := 0
	blk, err := db.GetTransactionBlock(hash)
	if err != nil {
		return 0, err
	}
	last, err := db.GetLastBlock()
	if err != nil {
		return 0, err
	}
	for last != nil {
		path += 1
		if bytes.Compare(last.Hash(), blk.Hash()) == 0 {
			break
		}
		last, err = db.GetBlock(last.PreviousBlockHash)
		if err != nil {
			return 0, err
		}
	}

	if path != 0 && bytes.Compare(last.Hash(), blk.Hash()) == 0 {
		return path, nil
	} else {
		return 0, nil
	}
}

func (db *T) GetPreviousTransactionHashForPublicKey(publicKey *ecdsa.PublicKey) (types.Hash, error) {
	enc, err := keys.EncodeECDSAPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("blocks"))
	if bucket == nil {
		return nil, nil
	}
	txHash := bucket.Get(append([]byte("<"), enc...))
	if txHash == nil {
		return nil, nil
	}
	return txHash, nil
}

func (db *T) GetNextTransactionHash(hash []byte) (types.Hash, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("blocks"))
	if bucket == nil {
		return nil, nil
	}
	txHash := bucket.Get(append([]byte(">"), hash...))
	if txHash == nil {
		return types.EmptyHash(), nil
	}
	return txHash, nil
}
