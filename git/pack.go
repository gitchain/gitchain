package git

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gitchain/gitchain/util"
)

type Delta struct {
	Hash  []byte
	Delta []byte
}

type Packfile struct {
	Version  uint32
	Objects  []Object
	Checksum []byte
	Deltas   []Delta
	offsets  map[int]int
	hashes   map[string]int
}

func (r *Packfile) ObjectByHash(hash []byte) Object {
	index, exists := r.hashes[string(hash)]
	if !exists {
		return nil
	}
	return r.Objects[index]
}

func (r *Packfile) ObjectByOffset(offset int) Object {
	index, exists := r.offsets[offset]
	if !exists {
		return nil
	}
	return r.Objects[index]
}

func (r *Packfile) PutObject(o Object) {
	r.Objects = append(r.Objects, o)
	r.hashes[string(o.Hash())] = len(r.Objects) - 1
}

func readMSBEncodedSize(reader io.Reader, initialOffset uint) uint64 {
	var b byte
	var sz uint64
	shift := initialOffset
	sz = 0
	for {
		binary.Read(reader, binary.BigEndian, &b)
		sz += (uint64(b) &^ 0x80) << shift
		shift += 7
		if (b & 0x80) == 0 {
			break
		}
	}
	return sz
}

func inflate(reader io.Reader, sz int) ([]byte, error) {
	zr, err := zlib.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("error opening packfile's object zlib: %v", err)
	}
	buf := make([]byte, sz)

	n, err := zr.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != sz {
		return nil, fmt.Errorf("inflated size mismatch, expected %d, got %d", sz, n)
	}

	zr.Close()
	return buf, nil
}

func readEntry(packfile *Packfile, reader flate.Reader, offset int) error {
	var b, typ uint8
	var sz uint64
	binary.Read(reader, binary.BigEndian, &b)
	typ = (b &^ 0x8f) >> 4
	sz = uint64(b &^ 0xf0)
	switch typ {
	case OBJ_REF_DELTA:
		if (b & 0x80) != 0 {
			sz += readMSBEncodedSize(reader, 4)
		}
		ref := make([]byte, 20)
		reader.Read(ref)

		buf, err := inflate(reader, int(sz))
		if err != nil {
			return err
		}

		referenced := packfile.ObjectByHash(ref)
		if referenced == nil {
			packfile.Deltas = append(packfile.Deltas, Delta{Hash: ref, Delta: buf})
		} else {
			patched := PatchDelta(referenced.Bytes(), buf)
			if patched == nil {
				return fmt.Errorf("error while patching %x", ref)
			}
			newObject := referenced.New()
			err = newObject.SetBytes(patched)
			if err != nil {
				return err
			}
			packfile.PutObject(newObject)
		}
	case OBJ_OFS_DELTA:
		if (b & 0x80) != 0 {
			sz += readMSBEncodedSize(reader, 4)
		}

		// read negative offset
		binary.Read(reader, binary.BigEndian, &b)
		var noffset int = int(b & 0x7f)
		for (b & 0x80) != 0 {
			noffset += 1
			binary.Read(reader, binary.BigEndian, &b)
			noffset = (noffset << 7) + int(b&0x7f)
		}

		buf, err := inflate(reader, int(sz))
		if err != nil {
			return err
		}
		referenced := packfile.ObjectByOffset(offset - noffset)
		if referenced == nil {
			return fmt.Errorf("can't find a pack entry at %d", offset-noffset)
		} else {
			patched := PatchDelta(referenced.Bytes(), buf)
			if patched == nil {
				return fmt.Errorf("error while patching %x", referenced.Hash())
			}
			newObject := referenced.New()
			err = newObject.SetBytes(patched)
			if err != nil {
				return err
			}
			packfile.PutObject(newObject)
		}
	case OBJ_COMMIT, OBJ_TREE, OBJ_BLOB, OBJ_TAG:
		if (b & 0x80) != 0 {
			sz += readMSBEncodedSize(reader, 4)
		}
		buf, err := inflate(reader, int(sz))
		if err != nil {
			return err
		}
		var obj Object
		switch typ {
		case OBJ_COMMIT:
			obj = &Commit{}
		case OBJ_TREE:
			obj = &Tree{}
		case OBJ_BLOB:
			obj = &Blob{}
		case OBJ_TAG:
			obj = &Tag{}
		}
		obj.SetBytes(buf)
		packfile.PutObject(obj)
	default:
		return fmt.Errorf("Invalid git object tag %03b", typ)
	}
	return nil
}

func ReadPackfile(r io.Reader) (*Packfile, error) {
	buffer, err := ioutil.ReadAll(r)
	contentChecksum := util.SHA160(buffer[0 : len(buffer)-20])
	r = bytes.NewBuffer(buffer)

	magic := make([]byte, 4)
	r.Read(magic)
	if bytes.Compare(magic, []byte("PACK")) != 0 {
		return nil, errors.New("not a packfile")
	}
	packfile := &Packfile{offsets: make(map[int]int), hashes: make(map[string]int)}

	var objects uint32
	binary.Read(r, binary.BigEndian, &packfile.Version)
	binary.Read(r, binary.BigEndian, &objects)

	content, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}
	offset := 12

	for i := 0; i < int(objects); i++ {
		peReader := &packEntryReader{reader: bytes.NewBuffer(content)}
		err := readEntry(packfile, peReader, offset)
		if err != nil {
			return packfile, err
		}
		packfile.offsets[offset] = len(packfile.Objects) - 1

		offset += peReader.Counter + 4
		content = content[peReader.Counter+4:]

	}

	var unresolvedDeltas []Delta
	for i := range packfile.Deltas {
		ref := packfile.ObjectByHash(packfile.Deltas[i].Hash)
		if ref == nil {
			unresolvedDeltas = append(unresolvedDeltas, packfile.Deltas[i])
		} else {
			patched := PatchDelta(ref.Bytes(), packfile.Deltas[i].Delta)
			newObject := ref.New()
			err = newObject.SetBytes(patched)
			if err != nil {
				return packfile, err
			}
			packfile.Objects = append(packfile.Objects, newObject)
		}
	}
	packfile.Deltas = unresolvedDeltas

	packfile.Checksum = make([]byte, 20)
	bytes.NewBuffer(content).Read(packfile.Checksum)

	if bytes.Compare(contentChecksum, packfile.Checksum) != 0 {
		return packfile, errors.New(fmt.Sprintf("checksum mismatch: expected %x got %x",
			packfile.Checksum, contentChecksum))
	}

	return packfile, nil
}

// This byte-counting hack is here to work around the fact that both zlib
// and flate use bufio and are very eager to read more data than they need.
// The counter in this reader allows us to know the length of the header +
// packed data read and therefore readjust the offset
type packEntryReader struct {
	Counter int
	reader  io.Reader
}

func (r *packEntryReader) Read(p []byte) (int, error) {
	r.Counter += (len(p))
	return r.reader.Read(p)
}

func (r *packEntryReader) ReadByte() (byte, error) {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}
