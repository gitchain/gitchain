package db

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/types"
)

func (db *T) PutBlock(b *block.Block, last bool) (e error) {
	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, e := dbtx.CreateBucketIfNotExists([]byte("blocks"))
		if e != nil {
			return false
		}
		encoded, e := b.Encode()
		if e != nil {
			return false
		}
		e = bucket.Put(b.Hash(), encoded)
		if e != nil {
			return false
		}
		// store a reference to this block as a next block
		e = bucket.Put(append(b.PreviousBlockHash, []byte("+")...), b.Hash())
		if e != nil {
			return false
		}

		for i := range b.Transactions {
			// transaction -> block mapping
			e = bucket.Put(append([]byte("T"), b.Transactions[i].Hash()...), b.Hash())
			if e != nil {
				return false
			}
			// link next transactions
			e = bucket.Put(append([]byte(">"), b.Transactions[i].PreviousEnvelopeHash...), b.Transactions[i].Hash())
			if e != nil {
				return false
			}
			// update "unspendable key"
			e = bucket.Delete(append([]byte("<"), b.Transactions[i].PublicKey...))
			if e != nil {
				return false
			}
			// update "spendable key", has to happen after updating the "unspendable" one as it might be the same one
			e = bucket.Put(append([]byte("<"), b.Transactions[i].NextPublicKey...), b.Transactions[i].Hash())
			if e != nil {
				return false
			}
		}

		if last {
			if e = bucket.Put([]byte("last"), b.Hash()); e != nil {
				return false
			}
		}

		return true
	})
	return
}

func (db *T) GetBlock(hash []byte) (blk *block.Block, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("blocks"))
		if bucket == nil {
			e = errors.New("blocks bucket does not exist")
			return
		}
		b := bucket.Get(hash)
		if b == nil {
			e = errors.New("block not found")
			return
		}
		blk, e = block.Decode(b)
	})
	return
}

func (db *T) GetLastBlock() (blk *block.Block, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("blocks"))
		if bucket == nil {
			return
		}
		last := bucket.Get([]byte("last"))
		if last == nil {
			return
		}
		b := bucket.Get(last)
		if b == nil {
			e = errors.New("block not found")
			return
		}
		blk, e = block.Decode(b)
	})
	return
}

func (db *T) GetNextBlock(hash types.Hash) (blk *block.Block, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("blocks"))
		if bucket == nil {
			return
		}
		next := bucket.Get(append(hash, []byte("+")...))
		if next == nil {
			return
		}
		b := bucket.Get(next)
		if b == nil {
			e = errors.New("block not found")
			return
		}
		blk, e = block.Decode(b)
	})
	return
}
