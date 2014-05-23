package net

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/gob"
	"math/big"

	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/util"
)

type KeyAuth struct {
	key *ecdsa.PrivateKey
}

func newKeyAuth() (ka *KeyAuth, e error) {
	pk, e := keys.GenerateECDSA()
	ka = &KeyAuth{key: pk}
	return
}

func (a *KeyAuth) Marshal() []byte {
	key, err := keys.EncodeECDSAPublicKey(&a.key.PublicKey)
	if err != nil {
		panic(err) // TODO: better error handling
	}
	sigR, sigS, err := ecdsa.Sign(rand.Reader, a.key, util.SHA256(key))
	if err != nil {
		panic(err) // TODO: better error handling
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode([][]byte{key, sigR.Bytes(), sigS.Bytes()})
	return buf.Bytes()
}

func (*KeyAuth) Valid(b []byte) bool {
	var data [][]byte
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	dec.Decode(&data)
	publicKey, err := keys.DecodeECDSAPublicKey(data[0])
	if err != nil {
		panic(err) // TODO: better error handling
	}
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(data[1])
	s.SetBytes(data[2])
	return ecdsa.Verify(publicKey, util.SHA256(data[0]), r, s)
}
