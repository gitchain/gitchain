package db

import (
	"errors"

	"../block"
	"../transaction"
	"github.com/boltdb/bolt"
)

type T struct {
	Path string
	DB   *bolt.DB
}

func NewDB(path string) (*T, error) {
	db, err := bolt.Open(path, 0666)
	return &T{Path: path, DB: db}, err
}

func (db *T) PutTransaction(txn transaction.T) error {
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
	encodedTx, err := txn.Encode()
	if err != nil {
		return err
	}
	err = bucket.Put(txn.Hash(), encodedTx)
	if err != nil {
		return err
	}
	success = true
	return nil
}

func (db *T) GetTransaction(id []byte) (transaction.T, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("transactions"))
	if bucket == nil {
		return nil, errors.New("transactions bucket does not exist")
	}
	txn := bucket.Get(id)
	if txn == nil {
		return nil, errors.New("transaction not found")
	}
	decodedTx, err := transaction.Decode(txn)
	if err != nil {
		return decodedTx, err
	}
	return decodedTx, nil
}

func (db *T) PutKey(alias string, key []byte) error {
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
	bucket, err := dbtx.CreateBucketIfNotExists([]byte("keys"))
	if err != nil {
		return err
	}
	err = bucket.Put([]byte(alias), key)
	if err != nil {
		return err
	}
	success = true
	return nil
}

func (db *T) GetKey(alias string) []byte {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil
	}
	bucket := dbtx.Bucket([]byte("keys"))
	if bucket == nil {
		return nil
	}
	return bucket.Get([]byte(alias))
}

func (db *T) ListKeys() []string {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil
	}
	bucket := dbtx.Bucket([]byte("keys"))
	if bucket == nil {
		return nil
	}
	keys := make([]string, 0)
	bucket.ForEach(func(k, v []byte) error {
		keys = append(keys, string(k))
		return nil
	})
	return keys
}

func (db *T) PutBlock(b *block.Block, last bool) error {
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
	bucket, err := dbtx.CreateBucketIfNotExists([]byte("blocks"))
	if err != nil {
		return err
	}
	encoded, err := b.Encode()
	if err != nil {
		return err
	}
	err = bucket.Put(b.Hash(), encoded)
	if err != nil {
		return err
	}
	if last {
		bucket.Put([]byte("last"), b.Hash())
	}
	success = true
	return nil
}

func (db *T) GetBlock(hash []byte) (*block.Block, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("blocks"))
	if bucket == nil {
		return nil, errors.New("blocks bucket does not exist")
	}
	b := bucket.Get(hash)
	if b == nil {
		return nil, errors.New("block not found")
	}
	decodedBlock, err := block.Decode(b)
	if err != nil {
		return decodedBlock, err
	}
	return decodedBlock, nil
}

func (db *T) GetLastBlock() (*block.Block, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("blocks"))
	if bucket == nil {
		return nil, errors.New("blocks bucket does not exist")
	}
	last := bucket.Get([]byte("last"))
	if last == nil {
		return nil, errors.New("no last block not found")
	}
	b := bucket.Get(last)
	if b == nil {
		return nil, errors.New("block not found")
	}
	decodedBlock, err := block.Decode(b)
	if err != nil {
		return decodedBlock, err
	}
	return decodedBlock, nil
}
