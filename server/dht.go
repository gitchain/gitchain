package server

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/util"
	"github.com/gitchain/wendy"
)

type Credentials struct{}

func (*Credentials) Marshal() []byte {
	return []byte{}
}

func (*Credentials) Valid([]byte) bool {
	return true
}

type GitchainApp struct {
}

func (app *GitchainApp) OnError(err error) {
	panic(err.Error())
}

func (app *GitchainApp) OnDeliver(msg wendy.Message) {
	fmt.Println("Received message: ", msg)
}

func (app *GitchainApp) OnForward(msg *wendy.Message, next wendy.NodeID) bool {
	log.Printf("Forwarding message %s to Node %s.", msg.Key, next)
	return true // return false if you don't want the message forwarded
}

func (app *GitchainApp) OnNewLeaves(leaves []*wendy.Node) {
	log.Println("Leaf set changed: ", leaves)
}

func (app *GitchainApp) OnNodeJoin(node wendy.Node) {
	log.Println("Node joined: ", node.ID)
}

func (app *GitchainApp) OnNodeExit(node wendy.Node) {
	log.Println("Node left: ", node.ID)
}

func (app *GitchainApp) OnHeartbeat(node wendy.Node) {
	log.Println("Received heartbeat from ", node.ID)
}

func DHTServer(srv *T) {
	ch := make(chan string)
	_, err := router.PermanentSubscribe("/dht/join", ch)

	id, err := wendy.NodeIDFromBytes(util.SHA160([]byte(srv.NetHostname)))
	if err != nil {
		log.Printf("Error preparing node ID: %v", err)
		os.Exit(0)
	}
	hostname := strings.Split(srv.NetHostname, ":")[0]
	node := wendy.NewNode(id, "127.0.0.1", hostname, "localhost", srv.NetPort)

	cluster := wendy.NewCluster(node, &Credentials{})
	cluster.SetLogLevel(wendy.LogLevelDebug)
	cluster.RegisterCallback(&GitchainApp{})
	go cluster.Listen()
loop:
	select {
	case existing := <-ch:
		log.Printf("Received a request to join the cluster at %s", existing)

		addr := strings.Split(existing, ":")
		port := 31000
		if len(addr) == 2 {
			port, err = strconv.Atoi(addr[1])
			if err != nil {
				fmt.Printf("Invalid port in %s: %v, ignoring the join request", addr[1], err)
				goto loop
			}
		}
		err = cluster.Join(addr[0], port)

		if err != nil {
			log.Printf("Error while joining cluster at %s: %v, becoming a disconnected node", existing, err)
			goto loop
		}
		log.Printf("Join request has been sent")

	}
	goto loop
}
