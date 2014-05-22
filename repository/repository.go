package repository

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"

	"github.com/gitchain/gitchain/types"
)

const (
	PENDING = 0
	ACTIVE  = 1
)

type Ref []byte

type T struct {
	Name             string
	Status           int
	NameAllocationTx types.Hash
}

func NewRepository(name string, status int, alloc types.Hash) *T {
	return &T{Name: name, Status: status, NameAllocationTx: alloc}
}

func (t *T) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&t)
	return buf.Bytes(), err
}

func Decode(encoded []byte) (*T, error) {
	var t T
	buf := bytes.NewBuffer(encoded)
	enc := gob.NewDecoder(buf)
	err := enc.Decode(&t)
	return &t, err
}

func EmptyRef() Ref {
	return make([]byte, 20)
}

func (r Ref) Equals(r1 Ref) bool {
	return bytes.Compare(r, r1) == 0
}

func (r Ref) String() string {
	return hex.EncodeToString(r)
}
