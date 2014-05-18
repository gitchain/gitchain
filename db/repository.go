package db

import (
	"github.com/gitchain/gitchain/repository"
)

func (db *T) PutRepository(repo *repository.T) error {
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
	bucket, err := dbtx.CreateBucketIfNotExists([]byte("repositories"))
	if err != nil {
		return err
	}

	encoded, err := repo.Encode()
	if err != nil {
		return err
	}

	err = bucket.Put([]byte(repo.Name), encoded)
	if err != nil {
		return err
	}

	if repo.Status == repository.PENDING {
		pendingBucket, err := dbtx.CreateBucketIfNotExists([]byte("pendingrepositories"))
		if err != nil {
			return err
		}
		err = pendingBucket.Put([]byte(repo.Name), []byte{})
		if err != nil {
			return err
		}
	}
	if repo.Status == repository.ACTIVE {
		pendingBucket, err := dbtx.CreateBucketIfNotExists([]byte("pendingrepositories"))
		if err != nil {
			return err
		}
		err = pendingBucket.Delete([]byte(repo.Name))
		if err != nil {
			return err
		}
	}

	success = true
	return nil
}

func (db *T) GetRepository(name string) (*repository.T, error) {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil, err
	}
	bucket := dbtx.Bucket([]byte("repositories"))
	if bucket == nil {
		return nil, nil // return no error because there were no repositories saved
	}
	b := bucket.Get([]byte(name))
	if b == nil {
		return nil, nil
	} else {
		decoded, err := repository.Decode(b)
		if err != nil {
			return nil, err
		}
		return decoded, nil
	}
}

func (db *T) ListRepositories() []string {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil
	}
	bucket := dbtx.Bucket([]byte("repositories"))
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

func (db *T) ListPendingRepositories() []string {
	dbtx, err := db.DB.Begin(false)
	defer dbtx.Rollback()
	if err != nil {
		return nil
	}
	bucket := dbtx.Bucket([]byte("pendingrepositories"))
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
