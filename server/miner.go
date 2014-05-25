package server

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/server/context"
)

type Miner struct {
	OriginalResponseChannel chan *block.Block `json:"-"`
	StartTime               time.Time
	NumTransactions         int
}

type MinerMap map[*block.Block]Miner

func (m MinerMap) MarshalJSON() ([]byte, error) {
	miners := make(map[string]interface{})
	for key, value := range m {
		miners[hex.EncodeToString(key.Hash())] = value
	}
	return json.Marshal(miners)
}

type MiningStatus struct {
	Miners      MinerMap
	BlocksMined int
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
	var response chan MiningStatus = make(chan MiningStatus)
	miningFactoryRequests <- MiningFactoryStatusRequest{ResponseChannel: response}
	return <-response
}

func MiningFactory(srv *context.T) {
	var status MiningStatus
	status.Miners = make(map[*block.Block]Miner)
	status.BlocksMined = 0
	minerCh := make(chan *block.Block)
loop:
	select {
	case minerMsg := <-minerCh:
		status.BlocksMined++
		status.Miners[minerMsg].OriginalResponseChannel <- minerMsg
		delete(status.Miners, minerMsg)
	case msg := <-miningFactoryRequests:
		switch msg.(type) {
		case MiningFactoryStatusRequest:
			sreq := msg.(MiningFactoryStatusRequest)
			sreq.ResponseChannel <- status
		case MiningFactoryInstantiationRequest:
			ireq := msg.(MiningFactoryInstantiationRequest)
			status.Miners[ireq.Block] = Miner{OriginalResponseChannel: ireq.ResponseChannel, StartTime: time.Now(), NumTransactions: len(ireq.Block.Transactions)}
			go ireq.Block.Mine(srv.Router, minerCh)
		}
	}
	goto loop
}
