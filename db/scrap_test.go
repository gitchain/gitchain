package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutGetDeleteScrap(t *testing.T) {
	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	err = db.PutScrap([]byte("hello"), []byte("world"))
	if err != nil {
		t.Errorf("error putting scrap: %v", err)
	}

	scrap, err := db.GetScrap([]byte("hello"))
	if err != nil {
		t.Errorf("error getting scrap: %v", err)
	}
	if scrap == nil {
		t.Errorf("error getting scrap `hello`")
	}
	assert.Equal(t, scrap, []byte("world"))

	err = db.DeleteScrap([]byte("hello"))
	if err != nil {
		t.Errorf("error deleting scrap: %v", err)
	}

	scrap, err = db.GetScrap([]byte("hello"))
	if err != nil {
		t.Errorf("error getting scrap: %v", err)
	}
	assert.Nil(t, scrap)

}
