package server

import (
	"time"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/server/context"
)

type Miner struct {
	signallingChannel chan *block.Block `json:"-"`
	responseChannel   chan *block.Block `json:"-"`
	Block             *block.Block
	StartTime         time.Time
}

type MiningStatus struct {
	Miners      []Miner
	BlocksMined int
}

func (s MiningStatus) AvailableMiners() (n int) {
	for i := range s.Miners {
		if s.Miners[i].Block == nil {
			n++
		}
	}
	return
}

type MiningFactoryStatusRequest struct {
	ResponseChannel chan MiningStatus
}

type MiningFactoryInstantiationRequest struct {
	Block           *block.Block
	ResponseChannel chan *block.Block
}

var miningFactoryRequests chan interface{} = make(chan interface{})

func GetMiningStatus() MiningStatus {
	response := make(chan MiningStatus)
	miningFactoryRequests <- &MiningFactoryStatusRequest{ResponseChannel: response}
	return <-response
}

func MiningFactory(srv *context.T) {
	log := srv.Log.New("cmp", "mining")
	var status MiningStatus
	n := srv.Config.Mining.Processes
	minedCh := make(chan *block.Block)
	for i := 0; i < n; i++ {
		ch := make(chan *block.Block)
		status.Miners = append(status.Miners, Miner{signallingChannel: ch})
		go block.Miner(ch, minedCh)
	}
loop:
	select {
	case block := <-minedCh:
		log.Debug("mined", "block", block)
		for i := range status.Miners {
			if status.Miners[i].Block == block {
				status.Miners[i].Block = nil
				status.Miners[i].StartTime = time.Unix(0, 0)
				status.Miners[i].responseChannel <- block
				status.BlocksMined++
				break
			}
		}
	case msg := <-miningFactoryRequests:
		switch msg.(type) {
		case *MiningFactoryStatusRequest:
			sreq := msg.(*MiningFactoryStatusRequest)
			sreq.ResponseChannel <- status
		case *MiningFactoryInstantiationRequest:
			ireq := msg.(*MiningFactoryInstantiationRequest)
			pickedMiner := -1
			for i := range status.Miners {
				if status.Miners[i].Block == nil {
					pickedMiner = i
					break
				}
			}
			if pickedMiner == -1 { // all miners are busy
				// pick the one with least transactions
				minTx := len(ireq.Block.Transactions)
				minIdx := -1
				for i := range status.Miners {
					if l := len(status.Miners[i].Block.Transactions); l < minTx {
						minTx = l
						minIdx = i
					}
				}
				if minIdx == -1 {
					// throw out this smaller block (for now?)
					goto loop
				} else {
					status.Miners[minIdx].signallingChannel <- nil
					pickedMiner = minIdx
				}
			}
			status.Miners[pickedMiner].responseChannel = ireq.ResponseChannel
			status.Miners[pickedMiner].Block = ireq.Block
			status.Miners[pickedMiner].StartTime = time.Now()
			status.Miners[pickedMiner].signallingChannel <- ireq.Block
		}
	}
	goto loop
}
