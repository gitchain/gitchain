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
)

type Packfile struct {
	Version  uint32
	Objects  []Object
	Checksum []byte
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
		return nil, errors.New(fmt.Sprintf("error opening packfile's object zlib: %v", err))
	}
	buf := make([]byte, sz)

	n, err := zr.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != sz {
		return nil, errors.New(fmt.Sprintf("inflated size mismatch, expected %d, got %d", sz, n))
	}

	zr.Close()
	return buf, nil
}

func readEntry(packfile *Packfile, reader flate.Reader) error {
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
		delta := make([]byte, 20)
		reader.Read(delta)

		_, err := inflate(reader, int(sz))
		if err != nil {
			return err
		}
		// packfile.Objects = append(packfile.Objects, buf)
	case OBJ_OFS_DELTA:
		if (b & 0x80) != 0 {
			sz += readMSBEncodedSize(reader, 4)
		}
		_, err := inflate(reader, int(sz))
		if err != nil {
			return err
		}
		// packfile.Objects = append(packfile.Objects, buf)
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
			obj = &Commit{Content: buf}
		case OBJ_TREE:
			obj = &Tree{Content: buf}
		case OBJ_BLOB:
			obj = &Blob{Content: buf}
		case OBJ_TAG:
			obj = &Tag{Content: buf}
		}
		packfile.Objects = append(packfile.Objects, obj)
	default:
		return errors.New(fmt.Sprintf("Invalid git object tag %03b", typ))
	}
	return nil
}

func ReadPackfile(r io.Reader) (*Packfile, error) {
	// bufreader := bufio.NewReader(r)

	magic := make([]byte, 4)
	r.Read(magic)
	if bytes.Compare(magic, []byte("PACK")) != 0 {
		return nil, errors.New("not a packfile")
	}
	packfile := &Packfile{}

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
		err := readEntry(packfile, peReader)
		if err != nil {
			return packfile, err
		}

		// retry
		// content1 := content[0 : peReader.Counter]
		// peReader1 := &packEntryReader{reader: bytes.NewBuffer(content1)}
		// err1 := readEntry(packfile, peReader1)
		// if err1 != nil {
		// 	return packfile, err1
		// }
		//

		offset += peReader.Counter + 4
		content = content[peReader.Counter+4:]

	}
	packfile.Checksum = make([]byte, 20)
	bytes.NewBuffer(content).Read(packfile.Checksum)
	return packfile, nil
}

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
