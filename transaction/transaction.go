package transaction

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"

	"github.com/conformal/fastsha256"
)

//// Interface

type T interface {
	Encode() ([]byte, error)
	Hash() []byte
}

func init() {
	gob.Register(&NameReservation{})
	gob.Register(&NameAllocation{})
	gob.Register(&NameDeallocation{})
}

func hash(t T) []byte {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(t)
	encoded := buf.Bytes()
	buf.Reset()
	binary.Write(buf, binary.BigEndian, fastsha256.Sum256(encoded))
	return buf.Bytes()
}

func encode(t T) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&t)
	return buf.Bytes(), err
}

func Decode(b []byte) (T, error) {
	var t T
	buf := bytes.NewBuffer(b)
	enc := gob.NewDecoder(buf)
	err := enc.Decode(&t)
	return t, err
}
