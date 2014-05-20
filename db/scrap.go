package db

import (
	"errors"

	"github.com/boltdb/bolt"
)

func (db *T) PutScrap(key, val []byte) (e error) {
	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, e := dbtx.CreateBucketIfNotExists([]byte("scraps"))
		if e != nil {
			return false
		}
		e = bucket.Put(key, val)
		return e == nil
	})
	return
}

func (db *T) GetScrap(key []byte) (b []byte, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("scraps"))
		if bucket == nil {
			e = errors.New("scraps bucket does not exist")
			return
		}
		b = bucket.Get(key)
	})
	return
}

func (db *T) DeleteScrap(key []byte) (e error) {
	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, e := dbtx.CreateBucketIfNotExists([]byte("scraps"))
		if e != nil {
			return false
		}
		e = bucket.Delete(key)
		return e == nil
	})
	return
}
