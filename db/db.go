package db

import (
	"errors"
	"gitchain/transaction"

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
	err = bucket.Put(txn.Id(), encodedTx)
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
