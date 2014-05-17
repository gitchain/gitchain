package db

import (
	"crypto/ecdsa"
	"github.com/gitchain/gitchain/keys"
)

func (db *T) PutKey(alias string, key *ecdsa.PrivateKey, main bool) error {
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

	encodedKey, err := keys.EncodeECDSAPrivateKey(key)
	if err != nil {
		return err
	}

	err = bucket.Put([]byte(alias), encodedKey)

	if err != nil {
		return err
	}
	if main {
		err = bucket.Put([]byte("main"), []byte(alias))
		if err != nil {
			return err
		}
	}
	success = true
	return nil
}

func (db *T) GetKey(alias string) (*ecdsa.PrivateKey, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("keys"))
	if bucket == nil {
		return nil, nil // return no error because there were no keys saved
	}
	key := bucket.Get([]byte(alias))
	if key == nil {
		return nil, nil
	} else {
		decodedKey, err := keys.DecodeECDSAPrivateKey(key)
		if err != nil {
			return nil, err
		}
		return decodedKey, nil
	}
}

func (db *T) GetMainKey() (*ecdsa.PrivateKey, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("keys"))
	if bucket == nil {
		return nil, nil // return no error because there were no keys saved
	}
	main := bucket.Get([]byte("main"))
	var key []byte
	if main == nil {
		keys := db.ListKeys()
		if len(keys) == 0 {
			return nil, nil
		} else {
			key = bucket.Get([]byte(keys[0]))
		}
	} else {
		key = bucket.Get(main)
	}
	if key == nil {
		return nil, nil
	} else {
		decodedKey, err := keys.DecodeECDSAPrivateKey(key)
		if err != nil {
			return nil, err
		}
		return decodedKey, nil
	}
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
