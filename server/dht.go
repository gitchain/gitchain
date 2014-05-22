package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
	"github.com/gitchain/gitchain/util"
	"github.com/gitchain/wendy"
)

const (
	MSG_BROADCAST   byte = 0x80
	MSG_REGULAR     byte = 0x40
	MSG_TRANSACTION byte = 0x01
)

type Credentials struct{}

func (*Credentials) Marshal() []byte {
	return []byte{}
}

func (*Credentials) Valid([]byte) bool {
	return true
}

type GitchainApp struct {
	cluster *wendy.Cluster
}

func (app *GitchainApp) OnError(err error) {
	panic(err.Error())
}

func (app *GitchainApp) OnDeliver(msg wendy.Message) {
	var err error
	if msg.Purpose&MSG_BROADCAST != 0 {
		var envelope broadcastEnvelope
		dec := gob.NewDecoder(bytes.NewBuffer(msg.Value))
		dec.Decode(&envelope)
		if err != nil {
			log.Printf("Error while decoding message: %v", err)
		} else {
			if msg.Purpose&MSG_TRANSACTION != 0 {
				var txne *transaction.Envelope
				if txne, err = transaction.DecodeEnvelope(envelope.Content); err != nil {
					log.Printf("Error while decoding transaction: %v", err)
				} else {
					router.Send("/transaction", make(chan *transaction.Envelope), txne)
				}
			}
			var newLimit wendy.NodeID
			nodes := app.cluster.RoutingTableNodes()
			for i := range nodes[0 : len(nodes)-2] {
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				if nodes[i].ID.Less(envelope.Limit) {
					if nodes[i+1].ID.Less(envelope.Limit) {
						newLimit = nodes[i+1].ID
					} else {
						newLimit = envelope.Limit
					}
					if err = enc.Encode(broadcastEnvelope{Content: envelope.Content, Limit: newLimit}); err != nil {
						return
					}
					wmsg := app.cluster.NewMessage(msg.Purpose, nodes[i].ID, buf.Bytes())
					if err = app.cluster.Send(wmsg); err != nil {
						log.Printf("Error sending message: %v", err)
					}
				} else {
					break
				}
			}
			if nodes[len(nodes)-1].ID.Less(envelope.Limit) {
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				if err = enc.Encode(broadcastEnvelope{Content: envelope.Content, Limit: app.cluster.ID()}); err != nil {
					return
				}
				wmsg := app.cluster.NewMessage(msg.Purpose, nodes[len(nodes)-1].ID, buf.Bytes())
				if err = app.cluster.Send(wmsg); err != nil {
					log.Printf("Error sending message: %v", err)
				}
			}

		}
	}
}

func (app *GitchainApp) OnForward(msg *wendy.Message, next wendy.NodeID) bool {
	return true
}

func (app *GitchainApp) OnNewLeaves(leaves []*wendy.Node) {
}

func (app *GitchainApp) OnNodeJoin(node wendy.Node) {
	log.Println("Node joined: ", node.ID)
}

func (app *GitchainApp) OnNodeExit(node wendy.Node) {
	log.Println("Node left: ", node.ID)
}

func (app *GitchainApp) OnHeartbeat(node wendy.Node) {
}

func DHTServer(srv *T) {
	ch := make(chan string)
	_, err := router.PermanentSubscribe("/dht/join", ch)
	tch := make(chan *transaction.Envelope)
	_, err = router.PermanentSubscribe("/transaction/mem", tch)

	id, err := wendy.NodeIDFromBytes(util.SHA160([]byte(srv.NetHostname)))
	if err != nil {
		log.Printf("Error preparing node ID: %v", err)
		os.Exit(0)
	}
	hostname := strings.Split(srv.NetHostname, ":")[0]
	node := wendy.NewNode(id, "127.0.0.1", hostname, "localhost", srv.NetPort)

	cluster := wendy.NewCluster(node, &Credentials{})
	cluster.SetLogLevel(wendy.LogLevelError)
	cluster.RegisterCallback(&GitchainApp{cluster: cluster})
	go cluster.Listen()
	defer cluster.Stop()
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
	case txe := <-tch:
		if err = broadcast(cluster, txe, MSG_TRANSACTION); err != nil {
			log.Println(err)
		}
	}
	goto loop
}

type HashableEncodable interface {
	Hash() types.Hash
	Encode() ([]byte, error)
}

type broadcastEnvelope struct {
	Content []byte
	Limit   wendy.NodeID
}

func init() {
	gob.Register(broadcastEnvelope{})
}

func broadcast(c *wendy.Cluster, msg HashableEncodable, purpose byte) (err error) {
	purpose = purpose | MSG_BROADCAST
	nodes := c.RoutingTableNodes()
	var b []byte
	var limit wendy.NodeID
	for i := range nodes {
		if b, err = msg.Encode(); err != nil {
			return
		}

		if i == len(nodes)-1 {
			limit = c.ID()
		} else {
			limit = nodes[i+1].ID
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err = enc.Encode(broadcastEnvelope{Content: b, Limit: limit}); err != nil {
			return
		}

		wmsg := c.NewMessage(purpose, nodes[i].ID, buf.Bytes())
		if err = c.Send(wmsg); err != nil {
			return
		}
	}
	return nil
}
