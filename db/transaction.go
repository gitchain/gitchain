package db

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/transaction"
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

func (db *T) GetPreviousEnvelopeHashForPublicKey(publicKey *ecdsa.PublicKey) (types.Hash, error) {
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

func (db *T) PutTransaction(tx *transaction.Envelope) error {
	dbtx, err := db.DB.Begin(true)
	success := false
	defer func() {
		if success {
			dbtx.Commit()
		} else {
			dbtx.Rollback()
		}
	}()
	if err != nil {
		return err
	}
	bucket, err := dbtx.CreateBucketIfNotExists([]byte("transactions"))
	if err != nil {
		return err
	}
	encoded, err := tx.Encode()
	if err != nil {
		return err
	}
	err = bucket.Put(tx.Hash(), encoded)
	if err != nil {
		return err
	}
	success = true
	return nil
}

func (db *T) GetTransaction(hash types.Hash) (*transaction.Envelope, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("transactions"))
	if bucket == nil {
		return nil, errors.New("transactions bucket does not exist")
	}
	b := bucket.Get(hash)
	if b == nil {
		return nil, nil
	}
	decodedTx, err := transaction.DecodeEnvelope(b)
	if err != nil {
		return decodedTx, err
	}
	return decodedTx, nil
}

func (db *T) DeleteTransaction(hash types.Hash) error {
	dbtx, err := db.DB.Begin(true)
	success := false
	defer func() {
		if success {
			dbtx.Commit()
		} else {
			dbtx.Rollback()
		}
	}()
	if err != nil {
		return err
	}
	bucket, err := dbtx.CreateBucketIfNotExists([]byte("transactions"))
	if err != nil {
		return err
	}
	err = bucket.Delete(hash)
	if err != nil {
		return err
	}
	success = true
	return nil
}
