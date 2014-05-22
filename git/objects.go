package git

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gitchain/gitchain/util"
)

type Object interface {
	Hash() []byte
	Bytes() []byte
	SetBytes([]byte)
	New() Object
	Type() string
}

const (
	OBJ_COMMIT    = 1
	OBJ_TREE      = 2
	OBJ_BLOB      = 3
	OBJ_TAG       = 4
	OBJ_OFS_DELTA = 6
	OBJ_REF_DELTA = 7
)

func ObjectToBytes(o Object) []byte {
	return append(append([]byte(fmt.Sprintf("%s %d", o.Type(), len(o.Bytes()))), 0), o.Bytes()...)
}

type Commit struct {
	Content []byte
}

func (o *Commit) Type() string {
	return "commit"
}

func (o *Commit) Hash() []byte {
	return util.SHA160(ObjectToBytes(o))
}

func (o *Commit) SetBytes(b []byte) {
	o.Content = b
}

func (o *Commit) Bytes() []byte {
	return o.Content
}

func (o *Commit) New() Object {
	return &Commit{}
}

func (o *Commit) String() string {
	return fmt.Sprintf("commit %x", o.Hash())
}

type Tree struct {
	Content []byte
}

func (o *Tree) Type() string {
	return "tree"
}

func (o *Tree) Hash() []byte {
	return util.SHA160(ObjectToBytes(o))
}

func (o *Tree) SetBytes(b []byte) {
	o.Content = b
}

func (o *Tree) Bytes() []byte {
	return o.Content
}

func (o *Tree) New() Object {
	return &Tree{}
}

func (o *Tree) String() string {
	return fmt.Sprintf("tree %x", o.Hash())
}

type Blob struct {
	Content []byte
}

func (o *Blob) Type() string {
	return "blob"
}

func (o *Blob) Hash() []byte {
	return util.SHA160(ObjectToBytes(o))
}

func (o *Blob) Bytes() []byte {
	return o.Content
}

func (o *Blob) SetBytes(b []byte) {
	o.Content = b
}

func (o *Blob) New() Object {
	return &Blob{}
}

func (o *Blob) String() string {
	return fmt.Sprintf("blob %x", o.Hash())
}

type Tag struct {
	Content []byte
}

func (o *Tag) Type() string {
	return "tag"
}

func (o *Tag) Hash() []byte {
	return util.SHA160(ObjectToBytes(o))
}

func (o *Tag) Bytes() []byte {
	return o.Content
}

func (o *Tag) SetBytes(b []byte) {
	o.Content = b
}

func (o *Tag) New() Object {
	return &Tag{}
}

func (o *Tag) String() string {
	return fmt.Sprintf("tag %x", o.Hash())
}

func WriteObject(o Object, dir string) (err error) {
	hash := []byte(hex.EncodeToString(o.Hash()))
	hd := hash[0:2]
	tl := hash[2:]
	err = os.MkdirAll(path.Join(dir, string(hd)), os.ModeDir|0700)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(dir, string(hd), string(tl)), o.Bytes(), 0600)
}

func DecodeObject(b []byte) (o Object) {
	split := bytes.Split(b, []byte{0})
	hdr := bytes.Split(split[0], []byte(" "))
	switch string(hdr[0]) {
	case "commit":
		o = &Commit{Content: split[1]}
	case "tree":
		o = &Tree{Content: split[1]}
	case "blob":
		o = &Blob{Content: split[1]}
	case "tag":
		o = &Tag{Content: split[1]}
	}
	return
}
