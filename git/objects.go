package git

import (
	"crypto/sha1"
	"fmt"
)

type Object interface {
	Hash() []byte
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

type Tree struct {
	Content []byte
}

func (o *Tree) Hash() []byte {
	result := sha1.Sum(append(append([]byte(fmt.Sprintf("tree %d", len(o.Content))), 0), o.Content...))
	return result[:]
}

type Blob struct {
	Content []byte
}

func (o *Blob) Hash() []byte {
	result := sha1.Sum(append(append([]byte(fmt.Sprintf("blob %d", len(o.Content))), 0), o.Content...))
	return result[:]
}

type Tag struct {
	Content []byte
}

func (o *Tag) Hash() []byte {
	result := sha1.Sum(append(append([]byte(fmt.Sprintf("tag %d", len(o.Content))), 0), o.Content...))
	return result[:]
}
