package db

import (
	"errors"
)

func (db *T) PutScrap(key, val []byte) error {
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
	bucket, err := dbtx.CreateBucketIfNotExists([]byte("scraps"))
	if err != nil {
		return err
	}
	err = bucket.Put(key, val)
	if err != nil {
		return err
	}

	success = true
	return nil
}

func (db *T) GetScrap(key []byte) ([]byte, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("scraps"))
	if bucket == nil {
		return nil, errors.New("scraps bucket does not exist")
	}
	b := bucket.Get(key)
	if b == nil {
		return nil, nil
	}
	return b, nil
}

func (db *T) DeleteScrap(key []byte) error {
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
	bucket, err := dbtx.CreateBucketIfNotExists([]byte("scraps"))
	if err != nil {
		return err
	}
	err = bucket.Delete(key)
	if err != nil {
		return err
	}

	success = true
	return nil
}
