package db

import (
	"bytes"
	"os"
	"sort"
	"testing"

	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/types"
	"github.com/gitchain/gitchain/util"
	"github.com/stretchr/testify/assert"
)

func TestPutGetRepository(t *testing.T) {

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	repo := repository.NewRepository("test", repository.PENDING, types.EmptyHash())
	err = db.PutRepository(repo)
	if err != nil {
		t.Errorf("error putting repository: %v", err)
	}
	repo1, err := db.GetRepository("test")
	if err != nil {
		t.Errorf("error getting repository: %v", err)
	}
	if repo1 == nil {
		t.Errorf("error getting repository `test'")
	}
	assert.Equal(t, repo, repo1)
}

func TestListRepository(t *testing.T) {

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	repo := repository.NewRepository("test", repository.PENDING, types.EmptyHash())
	err = db.PutRepository(repo)
	if err != nil {
		t.Errorf("error putting repository: %v", err)
	}

	repo1 := repository.NewRepository("hello_world", repository.ACTIVE, types.EmptyHash())
	err = db.PutRepository(repo1)
	if err != nil {
		t.Errorf("error putting repository: %v", err)
	}

	actualRepositories := db.ListRepositories()
	sort.Strings(actualRepositories)
	expectedRepositories := []string{"test", "hello_world"}
	sort.Strings(expectedRepositories)

	assert.Equal(t, actualRepositories, expectedRepositories)
}

func TestListPendingRepository(t *testing.T) {

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	repo := repository.NewRepository("test", repository.PENDING, types.EmptyHash())
	err = db.PutRepository(repo)
	if err != nil {
		t.Errorf("error putting repository: %v", err)
	}

	assert.Equal(t, db.ListPendingRepositories(), []string{"test"})

	repo1 := repository.NewRepository("hello_world", repository.ACTIVE, types.EmptyHash())
	err = db.PutRepository(repo1)
	if err != nil {
		t.Errorf("error putting repository: %v", err)
	}

	repo.Status = repository.ACTIVE
	err = db.PutRepository(repo)
	if err != nil {
		t.Errorf("error updating repository: %v", err)
	}

	assert.Equal(t, db.ListPendingRepositories(), []string{})

	actualRepositories := db.ListRepositories()
	sort.Strings(actualRepositories)
	expectedRepositories := []string{"test", "hello_world"}
	sort.Strings(expectedRepositories)

	assert.Equal(t, actualRepositories, expectedRepositories)
}

func TestPutGetRef(t *testing.T) {

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	repo := repository.NewRepository("myrepo", repository.ACTIVE, types.EmptyHash())
	err = db.PutRepository(repo)
	if err != nil {
		t.Errorf("error putting repository: %v", err)
	}

	// before the ref is set...
	ref0, err := db.GetRef("myrepo", "refs/heads/master")
	if err != nil {
		t.Errorf("error getting repository ref: %v", err)
	}
	assert.True(t, ref0.Equals(repository.EmptyRef()))

	ref := util.SHA160([]byte("random"))
	err = db.PutRef("myrepo", "refs/heads/master", ref)
	if err != nil {
		t.Errorf("error putting repository ref: %v", err)
	}
	ref1, err := db.GetRef("myrepo", "refs/heads/master")
	if err != nil {
		t.Errorf("error getting repository ref: %v", err)
	}
	if ref1 == nil {
		t.Errorf("error getting repository ref `refs/heads/master'")
	}
	assert.True(t, bytes.Compare(ref, ref1) == 0)
}

func TestListRefs(t *testing.T) {

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	repo := repository.NewRepository("myrepo", repository.ACTIVE, types.EmptyHash())
	err = db.PutRepository(repo)
	if err != nil {
		t.Errorf("error putting repository: %v", err)
	}

	ref := util.SHA160([]byte("random"))
	err = db.PutRef("myrepo", "refs/heads/master", ref)
	if err != nil {
		t.Errorf("error putting repository ref: %v", err)
	}
	err = db.PutRef("myrepo", "refs/heads/next", ref)
	if err != nil {
		t.Errorf("error putting repository ref: %v", err)
	}

	refs, err := db.ListRefs("myrepo")
	if err != nil {
		t.Errorf("error listing repository refs: %v", err)
	}

	assert.Equal(t, refs, []string{"refs/heads/master", "refs/heads/next"})
}
