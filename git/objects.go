package git

import (
	"crypto/sha1"
	"fmt"
)

type Object interface {
	Hash() []byte
	Bytes() []byte
	SetBytes([]byte)
	New() Object
}

const (
	OBJ_COMMIT    = 1
	OBJ_TREE      = 2
	OBJ_BLOB      = 3
	OBJ_TAG       = 4
	OBJ_OFS_DELTA = 6
	OBJ_REF_DELTA = 7
)

type Commit struct {
	Content []byte
}

func (o *Commit) Hash() []byte {
	result := sha1.Sum(append(append([]byte(fmt.Sprintf("commit %d", len(o.Content))), 0), o.Content...))
	return result[:]
}

func (o *Commit) Bytes() []byte {
	return o.Content
}

func (o *Commit) SetBytes(b []byte) {
	o.Content = b
}

func (o *Commit) New() Object {
	return &Commit{}
}

type Tree struct {
	Content []byte
}

func (o *Tree) Hash() []byte {
	result := sha1.Sum(append(append([]byte(fmt.Sprintf("tree %d", len(o.Content))), 0), o.Content...))
	return result[:]
}

func (o *Tree) Bytes() []byte {
	return o.Content
}

func (o *Tree) SetBytes(b []byte) {
	o.Content = b
}

func (o *Tree) New() Object {
	return &Tree{}
}

type Blob struct {
	Content []byte
}

func (o *Blob) Hash() []byte {
	result := sha1.Sum(append(append([]byte(fmt.Sprintf("blob %d", len(o.Content))), 0), o.Content...))
	return result[:]
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

type Tag struct {
	Content []byte
}

func (o *Tag) Hash() []byte {
	result := sha1.Sum(append(append([]byte(fmt.Sprintf("tag %d", len(o.Content))), 0), o.Content...))
	return result[:]
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
