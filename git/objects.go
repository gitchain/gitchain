package git

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gitchain/gitchain/util"
)

type Hash []byte

type Object interface {
	Hash() []byte
	Bytes() []byte
	SetBytes([]byte) error
	New() Object
	Type() string
}

func ObjectToBytes(o Object) []byte {
	return append(append([]byte(fmt.Sprintf("%s %d", o.Type(), len(o.Bytes()))), 0), o.Bytes()...)
}

type Commit struct {
	Content   []byte
	Tree      Hash
	Parents   []Hash
	Author    string
	Committer string
	Message   string
}

func (o *Commit) Type() string {
	return "commit"
}

func (o *Commit) Hash() []byte {
	return util.SHA160(ObjectToBytes(o))
}

func (o *Commit) SetBytes(b []byte) (err error) {
	o.Content = b
	lines := bytes.Split(b, []byte{'\n'})
	for i := range lines {
		if len(lines[i]) > 0 {
			split := bytes.SplitN(lines[i], []byte{' '}, 2)
			switch string(split[0]) {
			case "tree":
				o.Tree = make([]byte, 20)
				_, err = hex.Decode(o.Tree, split[1])
			case "parent":
				h := make([]byte, 20)
				_, err = hex.Decode(h, split[1])
				if err == nil {
					o.Parents = append(o.Parents, h)
				}
			case "author":
				o.Author = string(split[1])
			case "committer":
				o.Committer = string(split[1])
			}
			if err != nil {
				return
			}
		} else {
			o.Message = string(bytes.Join(append(lines[i+1:]), []byte{'\n'}))
			break
		}
	}
	return
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

type treeEntry struct {
	Mode string
	File string
	Hash Hash
}
type Tree struct {
	Content []byte
	Entries []treeEntry
}

func (o *Tree) Type() string {
	return "tree"
}

func (o *Tree) Hash() []byte {
	return util.SHA160(ObjectToBytes(o))
}

func (o *Tree) SetBytes(b []byte) (err error) {
	zr, e := zlib.NewReader(bytes.NewBuffer(b))
	if e == nil {
		defer zr.Close()
		b, err = ioutil.ReadAll(zr)
		if err != nil {
			return err
		}
	}

	o.Content = b
	body := b

	for {
		split := bytes.SplitN(body, []byte{0}, 2)
		split1 := bytes.SplitN(split[0], []byte{' '}, 2)
		o.Entries = append(o.Entries, treeEntry{
			Mode: string(split1[0]),
			File: string(split1[1]),
			Hash: split[1][0:20]})
		body = split[1][20:]
		if len(split[1]) == 20 {
			break
		}
	}
	return
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

func (o *Blob) SetBytes(b []byte) (err error) {
	o.Content = b
	return
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

func (o *Tag) SetBytes(b []byte) (err error) {
	o.Content = b
	return
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
	return ioutil.WriteFile(path.Join(dir, string(hd), string(tl)), ObjectToBytes(o), 0600)
}

func DecodeObject(b []byte) (o Object) {
	split := bytes.SplitN(b, []byte{0}, 2)
	hdr := bytes.Split(split[0], []byte(" "))
	switch string(hdr[0]) {
	case "commit":
		o = &Commit{}
	case "tree":
		o = &Tree{}
	case "blob":
		o = &Blob{}
	case "tag":
		o = &Tag{}
	}
	o.SetBytes(split[1])
	return
}
