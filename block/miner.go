package block

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"runtime"
	"time"

	"github.com/conformal/fastsha256"
	"github.com/tuxychandru/pubsub"
)

func (b *Block) Mine(router *pubsub.PubSub, c chan *Block) {
	lastch := router.Sub("/block/last")
	blockch := router.Sub("/block")
	defer router.Unsub(lastch)
	defer router.Unsub(blockch)

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
		case lasti := <-lastch:
			if last, ok := lasti.(*Block); ok {
				b.PreviousBlockHash = last.Hash()
				b.Timestamp = time.Now().UTC().Unix()
				goto loop
			}
		case ablocki := <-blockch:
			if ablock, ok := ablocki.(*Block); ok {

				for i := range ablock.Transactions {
					for j := range b.Transactions {
						if bytes.Compare(ablock.Transactions[i].Hash(), b.Transactions[j].Hash()) == 0 {
							b.Transactions = append(b.Transactions[0:j], b.Transactions[j+1:]...)
						}
					}
				}
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
				c <- b
				return
			}
			runtime.Gosched()
		}
	}
	// Update timestamp and restart the process
	b.Timestamp = time.Now().UTC().Unix()
	goto loop
}
