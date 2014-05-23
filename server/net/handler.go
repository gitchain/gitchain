package net

import (
	"bytes"
	"encoding/gob"
	"path"

	"github.com/gitchain/gitchain/git"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/server"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/wendy"
	"github.com/inconshreveable/log15"
)

type GitchainApp struct {
	cluster *wendy.Cluster
	log     log15.Logger
	srv     *server.T
}

func (app *GitchainApp) OnError(err error) {
	panic(err.Error())
}

func (app *GitchainApp) OnDeliver(msg wendy.Message) {
	log := app.log
	var err error
	if msg.Purpose&MSG_BROADCAST != 0 {
		log.Debug("received a broadcast")
		if msg.Sender.ID == app.cluster.ID() {
			log.Error("received own broadcast", "bugtrap", "true")
		}
		var envelope broadcastEnvelope
		dec := gob.NewDecoder(bytes.NewBuffer(msg.Value))
		dec.Decode(&envelope)
		if err != nil {
			log.Error("error while decoding an incoming message", "err", err)
		} else {
			var txe *transaction.Envelope
			if msg.Purpose&MSG_TRANSACTION != 0 {
				if txe, err = transaction.DecodeEnvelope(envelope.Content); err != nil {
					log.Error("error while decoding transaction", "err", err)
				} else {
					router.Send("/transaction", make(chan *transaction.Envelope), txe)
					log.Debug("announced transaction locally", "txn", txe)
				}
			}
			var newLimit wendy.NodeID
			nodes := app.cluster.RoutingTableNodes()
			if len(nodes) > 1 {
				for i := range nodes[0 : len(nodes)-1] {
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
						} else {
							log.Debug("forwarded transaction", "txn", txe)
						}
					} else {
						break
					}
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
				} else {
					log.Debug("forwarded transaction", "txn", txe)
				}
			}

		}
	} else {
		switch {
		case msg.Purpose&MSG_OBJECT != 0:
			obj := git.DecodeObject(msg.Value)
			err = git.WriteObject(obj, path.Join(app.srv.Config.General.DataPath, "objects"))
			if err != nil {
				log.Error("error while writing object", "obj", obj, "err", err)
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
