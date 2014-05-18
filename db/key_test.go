package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutGetListKey(t *testing.T) {
	key := generateECDSAKey(t)

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	// Before we do anything, try fetching the main key and make sure
	// there is none
	mainKey, err := db.GetMainKey()
	if err != nil {
		t.Errorf("error getting main key: %v", err)
	}
	assert.Nil(t, mainKey)
	assert.Equal(t, db.ListKeys(), []string(nil))

	err = db.PutKey("alias", key, false)
	if err != nil {
		t.Errorf("error putting key: %v", err)
	}
	key1, err := db.GetKey("alias")
	if err != nil {
		t.Errorf("error putting key: %v", err)
	}
	if key1 == nil {
		t.Errorf("no key was retrieved")
	}
	assert.Equal(t, key, key1)
	assert.Equal(t, db.ListKeys(), []string{"alias"})

	// Even though we did specify this key as non-main, it will still be
	// considered main as the first key
	key2, err := db.GetMainKey()
	if err != nil {
		t.Errorf("error getting main key: %v", err)
	}
	if key2 == nil {
		t.Errorf("there should be a main key anyway")
	}
	assert.Equal(t, key, key2)

	// Try adding another key that goes before (alphabetically)
	aaronkey := generateECDSAKey(t)

	err = db.PutKey("aaron", aaronkey, false)
	key3, err := db.GetMainKey()
	if err != nil {
		t.Errorf("error getting main key: %v", err)
	}
	if key3 == nil {
		t.Errorf("there should be an implicit main key")
	}
	assert.Equal(t, aaronkey, key3)

	// Try adding another key that goes after (alphabetically)
	betakey := generateECDSAKey(t)

	err = db.PutKey("beta", betakey, false)
	key31, err := db.GetMainKey()
	if err != nil {
		t.Errorf("error getting main key: %v", err)
	}
	if key31 == nil {
		t.Errorf("there should be an implicit main key")
	}
	assert.Equal(t, aaronkey, key31)

	// This proves that the first key by order, in absence of an explicitly
	// set main key, will be considered main

	// Try adding another key and setting it as a main key explicitly
	testkey := generateECDSAKey(t)

	err = db.PutKey("test", testkey, true)
	key4, err := db.GetMainKey()
	if err != nil {
		t.Errorf("error getting main key: %v", err)
	}
	if key4 == nil {
		t.Errorf("there should be an explicit main key")
	}
	assert.Equal(t, testkey, key4)

	// Now, try adding another key!
	charliekey := generateECDSAKey(t)

	err = db.PutKey("beta", charliekey, false)
	key41, err := db.GetMainKey()
	if err != nil {
		t.Errorf("error getting main key: %v", err)
	}
	if key41 == nil {
		t.Errorf("there should be an explicit main key")
	}
	assert.Equal(t, testkey, key41)

}
