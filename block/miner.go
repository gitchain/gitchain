package block

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"runtime"
	"time"

	"github.com/conformal/fastsha256"
)

func Miner(blockChan chan *Block, minedBlockChan chan *Block) {
	var b *Block
init:
	b = <-blockChan
	if b == nil {
		goto init
	}
mine:
	target := targetFromBits(b.Bits)
	i := big.NewInt(int64(0))
	var n uint32
	buf := bytes.NewBuffer([]byte{})
	buf.Grow(192)
	buf1 := bytes.NewBuffer([]byte{})
	buf1.Grow(32)
loop:
	for n = 0; n < 4294967295; n++ {
		select {
		case b = <-blockChan:
			if b == nil { // cancel mining
				goto init
			} else {
				goto mine
			}
		default:
			binary.Write(buf, binary.LittleEndian, b.PreviousBlockHash)
			binary.Write(buf, binary.LittleEndian, b.MerkleRootHash)
			binary.Write(buf, binary.LittleEndian, b.Version)
			binary.Write(buf, binary.LittleEndian, b.Timestamp)
			binary.Write(buf, binary.LittleEndian, b.Bits)
			binary.Write(buf, binary.LittleEndian, n) // current nonce
			hash := fastsha256.Sum256(buf.Bytes())
			binary.Write(buf1, binary.BigEndian, hash)
			hash = fastsha256.Sum256(buf1.Bytes())
			buf1.Reset()
			binary.Write(buf1, binary.BigEndian, hash)
			i.SetBytes(buf1.Bytes())
			buf1.Reset()
			buf.Reset()
			if i.Cmp(target) == -1 {
				b.Nonce = n
				minedBlockChan <- b
				goto init
			}
			runtime.Gosched()
		}
	}
	// Update timestamp and restart the process
	b.Timestamp = time.Now().UTC().Unix()
	goto loop
}
