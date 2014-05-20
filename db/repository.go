package db

import (
	"github.com/boltdb/bolt"
	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/types"
)

func (db *T) PutRepository(repo *repository.T) (e error) {
	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, e := dbtx.CreateBucketIfNotExists([]byte("repositories"))
		if e != nil {
			return false
		}

		encoded, e := repo.Encode()
		if e != nil {
			return false
		}

		e = bucket.Put([]byte(repo.Name), encoded)
		if e != nil {
			return false
		}

		if repo.Status == repository.PENDING {
			pendingBucket, e := dbtx.CreateBucketIfNotExists([]byte("pendingrepositories"))
			if e != nil {
				return false
			}
			e = pendingBucket.Put([]byte(repo.Name), []byte{})
			if e != nil {
				return false
			}
		}
		if repo.Status == repository.ACTIVE {
			pendingBucket, err := dbtx.CreateBucketIfNotExists([]byte("pendingrepositories"))
			if e != nil {
				return false
			}
			e = pendingBucket.Delete([]byte(repo.Name))
			if err != nil {
				return false
			}
		}
		return true
	})
	return
}

func (db *T) GetRepository(name string) (r *repository.T, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("repositories"))
		if bucket == nil {
			return
		}
		b := bucket.Get([]byte(name))
		if b == nil {
			return
		} else {
			r, e = repository.Decode(b)
		}
	})
	return
}

func (db *T) ListRepositories() (keys []string) {
	readable(nil, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("repositories"))
		if bucket == nil {
			return
		}
		keys = make([]string, 0)
		bucket.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return
}

func (db *T) ListPendingRepositories() (keys []string) {
	readable(nil, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("pendingrepositories"))
		if bucket == nil {
			return
		}
		keys = make([]string, 0)
		bucket.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return
}

func (db *T) PutRef(name, ref string, new types.Hash) (e error) {
	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, e := dbtx.CreateBucketIfNotExists([]byte("repositories"))
		if e != nil {
			return false
		}
		bucket, e = dbtx.CreateBucketIfNotExists(append([]byte("refs"), []byte(name)...))
		if e != nil {
			return false
		}

		e = bucket.Put([]byte(ref), new)
		return e == nil
	})
	return
}

func (db *T) GetRef(name, ref string) (h types.Hash, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("repositories"))
		if bucket == nil {
			h = types.EmptyHash() // return no error because there were no repositories saved
			return
		}
		bucket = dbtx.Bucket(append([]byte("refs"), []byte(name)...))
		if bucket == nil {
			h = types.EmptyHash() // return no error because there were no repositories saved
			return
		}
		h = bucket.Get([]byte(ref))
		if h == nil {
			h = types.EmptyHash()
		}
	})
	return
}

func (db *T) ListRefs(name string) (refs []string, e error) {
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("repositories"))
		if bucket == nil {
			refs = []string{} // return no error because there were no repositories saved
			return
		}
		bucket = dbtx.Bucket(append([]byte("refs"), []byte(name)...))
		if bucket == nil {
			refs = []string{} // return no error because there were no repositories saved
			return
		}
		refs = make([]string, 0)
		bucket.ForEach(func(k, v []byte) error {
			refs = append(refs, string(k))
			return nil
		})
	})
	return
}
