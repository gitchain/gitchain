package db

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
)

func (db *T) GetTransactionBlock(hash types.Hash) (b *block.Block, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("blocks"))
		if bucket != nil {
			blockHash := bucket.Get(append([]byte("T"), hash...))
			if blockHash == nil {
				e = fmt.Errorf("transaction %s isn't included in any block", hash)
				return
			}
			encodedBlock := bucket.Get(blockHash)
			b, e = block.Decode(encodedBlock)
		}
	})
	return
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

func (db *T) GetPreviousEnvelopeHashForPublicKey(publicKey *ecdsa.PublicKey) (h types.Hash, e error) {
	enc, e := keys.EncodeECDSAPublicKey(publicKey)
	if e != nil {
		return
	}
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("blocks"))
		if bucket != nil {
			h = bucket.Get(append([]byte("<"), enc...))
		}
	})
	return
}

func (db *T) GetNextTransactionHash(hash []byte) (h types.Hash, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("blocks"))
		if bucket != nil {
			h = bucket.Get(append([]byte(">"), hash...))
			if h == nil {
				h = types.EmptyHash()
			}
		}
	})
	return
}

func (db *T) PutTransaction(tx *transaction.Envelope) (e error) {
	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, e := dbtx.CreateBucketIfNotExists([]byte("transactions"))
		if e != nil {
			return false
		}
		encoded, e := tx.Encode()
		if e != nil {
			return false
		}
		e = bucket.Put(tx.Hash(), encoded)
		if e != nil {
			return false
		}
		return true
	})
	return
}

func (db *T) GetTransaction(hash types.Hash) (envelope *transaction.Envelope, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("transactions"))
		if bucket == nil {
			e = errors.New("transactions bucket does not exist")
			return
		}
		b := bucket.Get(hash)
		if b != nil {
			envelope, e = transaction.DecodeEnvelope(b)
		}
	})
	return
}

func (db *T) DeleteTransaction(hash types.Hash) (e error) {
	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, e := dbtx.CreateBucketIfNotExists([]byte("transactions"))
		if e != nil {
			return false
		}
		e = bucket.Delete(hash)
		if e != nil {
			return false
		}
		return true
	})
	return
}
