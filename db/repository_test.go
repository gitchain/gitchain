package db

import (
	"os"
	"sort"
	"testing"

	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/types"
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
