package server

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/gob"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
	"github.com/gitchain/gitchain/util"
	"github.com/gitchain/wendy"
	"github.com/inconshreveable/log15"
)

const (
	MSG_BROADCAST   byte = 0x80
	MSG_REGULAR     byte = 0x40
	MSG_TRANSACTION byte = 0x01
)

type KeyAuth struct {
	key *ecdsa.PrivateKey
}

func newKeyAuth() (ka *KeyAuth, e error) {
	pk, e := keys.GenerateECDSA()
	ka = &KeyAuth{key: pk}
	return
}

func (a *KeyAuth) Marshal() []byte {
	key, err := keys.EncodeECDSAPublicKey(&a.key.PublicKey)
	if err != nil {
		panic(err) // TODO: better error handling
	}
	sigR, sigS, err := ecdsa.Sign(rand.Reader, a.key, util.SHA256(key))
	if err != nil {
		panic(err) // TODO: better error handling
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode([][]byte{key, sigR.Bytes(), sigS.Bytes()})
	return buf.Bytes()
}

func (*KeyAuth) Valid(b []byte) bool {
	var data [][]byte
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	dec.Decode(&data)
	publicKey, err := keys.DecodeECDSAPublicKey(data[0])
	if err != nil {
		panic(err) // TODO: better error handling
	}
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(data[1])
	s.SetBytes(data[2])
	return ecdsa.Verify(publicKey, util.SHA256(data[0]), r, s)
}

type GitchainApp struct {
	cluster *wendy.Cluster
	log     log15.Logger
}

func (app *GitchainApp) OnError(err error) {
	panic(err.Error())
}

func (app *GitchainApp) OnDeliver(msg wendy.Message) {
	log := app.log
	var err error
	if msg.Purpose&MSG_BROADCAST != 0 {
		log.Debug("received a broadcast")
		var envelope broadcastEnvelope
		dec := gob.NewDecoder(bytes.NewBuffer(msg.Value))
		dec.Decode(&envelope)
		if err != nil {
			log.Error("error while decoding an incoming message", "err", err)
		} else {
			if msg.Purpose&MSG_TRANSACTION != 0 {
				var txne *transaction.Envelope
				if txne, err = transaction.DecodeEnvelope(envelope.Content); err != nil {
					log.Error("error while decoding transaction", "err", err)
				} else {
					router.Send("/transaction", make(chan *transaction.Envelope), txne)
					log.Debug("announced transaction locally", "txn", txne)
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
						log.Error("error sending message", "err", err)
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
					log.Error("error sending message", "err", err)
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
	app.log.Info("node joined", "node", node.ID, "addr", app.cluster.GetIP(node))
}

func (app *GitchainApp) OnNodeExit(node wendy.Node) {
	app.log.Info("node left", "node", node.ID)
}

func (app *GitchainApp) OnHeartbeat(node wendy.Node) {
}

func DHTServer(srv *T) {
	log := srv.Log.New("cmp", "dht")

	ch := make(chan string)
	_, err := router.PermanentSubscribe("/dht/join", ch)
	tch := make(chan *transaction.Envelope)
	_, err = router.PermanentSubscribe("/transaction/mem", tch)

	keyAuth, err := newKeyAuth()
	if err != nil {
		log.Crit("can't generate node key", "err", err)
		os.Exit(1)
	}

	id, err := wendy.NodeIDFromBytes(util.SHA256(keyAuth.Marshal()))
	log = log.New("own_node", id)

	if err != nil {
		log15.Crit("error preparing node id", "err", err)
		os.Exit(0)
	}

	hostname := strings.Split(srv.NetHostname, ":")[0]
	node := wendy.NewNode(id, "127.0.0.1", hostname, "localhost", srv.NetPort)

	cluster := wendy.NewCluster(node, keyAuth)
	cluster.SetLogLevel(wendy.LogLevelError)
	cluster.RegisterCallback(&GitchainApp{cluster: cluster, log: log.New()})
	go cluster.Listen()
	defer cluster.Stop()

	log.Info("node started")

loop:
	select {
	case existing := <-ch:
		log.Debug("received a request to join the cluster", "addr", existing)

		addr := strings.Split(existing, ":")
		port := 31000
		if len(addr) == 2 {
			port, err = strconv.Atoi(addr[1])
			if err != nil {
				log.Error("invalid port number", "addr", existing, "port", addr[1], "err", err)
				goto loop
			}
		}
		err = cluster.Join(addr[0], port)

		if err != nil {
			log.Error("can't join cluster", "addr", existing, "err", err)
			goto loop
		}
	case txe := <-tch:
		log.Debug("received transaction", "txn", txe)
		if err = broadcast(cluster, txe, MSG_TRANSACTION); err != nil {
			log.Error("error broadcasting a transaction message", "txn", txe)
		}
		log.Debug("broadcasted transaction", "txn", txe)
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
