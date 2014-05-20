package db

import (
	"crypto/ecdsa"

	"github.com/boltdb/bolt"
	"github.com/gitchain/gitchain/keys"
)

func (db *T) PutKey(alias string, key *ecdsa.PrivateKey, main bool) (e error) {
	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, e := dbtx.CreateBucketIfNotExists([]byte("keys"))
		if e != nil {
			return false
		}

		encodedKey, e := keys.EncodeECDSAPrivateKey(key)
		if e != nil {
			return false
		}

		e = bucket.Put([]byte(alias), encodedKey)

		if e != nil {
			return false
		}
		if main {
			e = bucket.Put([]byte("main"), []byte(alias))
			if e != nil {
				return false
			}
		}
		return true
	})
	return
}

func (db *T) GetKey(alias string) (k *ecdsa.PrivateKey, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("keys"))
		if bucket == nil {
			return // return no error because there were no keys saved
		}
		key := bucket.Get([]byte(alias))
		if key == nil {
			return
		} else {
			k, e = keys.DecodeECDSAPrivateKey(key)
		}
	})
	return
}

func (db *T) GetMainKey() (k *ecdsa.PrivateKey, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("keys"))
		if bucket == nil {
			return // return no error because there were no keys saved
		}
		main := bucket.Get([]byte("main"))
		var key []byte
		if main == nil {
			keys := db.ListKeys()
			if len(keys) == 0 {
				return
			} else {
				key = bucket.Get([]byte(keys[0]))
			}
		} else {
			key = bucket.Get(main)
		}
		if key == nil {
			return
		} else {
			k, e = keys.DecodeECDSAPrivateKey(key)
		}
	})
	return
}

func (db *T) ListKeys() (keys []string) {
	readable(nil, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("keys"))
		if bucket == nil {
			return
		}
		keys = make([]string, 0)
		bucket.ForEach(func(k, v []byte) error {
			if string(k) != "main" {
				keys = append(keys, string(k))
			}
			return nil
		})
	})
	return
}
