package net

import (
	"os"
	"strconv"
	"strings"

	"github.com/gitchain/gitchain/git"
	"github.com/gitchain/gitchain/server"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/util"
	"github.com/gitchain/wendy"
	"github.com/inconshreveable/log15"
)

func Server(srv *server.T) {
	log := srv.Log.New("cmp", "dht")
	ch := srv.Router.Sub("/dht/join")
	tch := srv.Router.Sub("/transaction/mem")
	och := srv.Router.Sub("/git/object")

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

	hostname := strings.Split(srv.Config.Network.Hostname, ":")[0]
	node := wendy.NewNode(id, "127.0.0.1", hostname, "localhost", srv.Config.Network.Port)

	cluster := wendy.NewCluster(node, keyAuth)
	cluster.SetLogLevel(wendy.LogLevelError)
	cluster.RegisterCallback(&GitchainApp{cluster: cluster, log: log.New(), srv: srv})
	go cluster.Listen()
	defer cluster.Stop()

	log.Info("node started")

	for i := range srv.Config.Network.Join {
		log.Info("scheduling a connection", "addr", srv.Config.Network.Join[i])
		srv.Router.Pub(srv.Config.Network.Join[i], "/dht/join")
	}

loop:
	select {
	case addri := <-ch:
		if addr, ok := addri.(string); ok {
			log.Debug("received a request to join the cluster", "addr", addr)

			addr := strings.Split(addr, ":")
			port := 31000
			if len(addr) == 2 {
				port, err = strconv.Atoi(addr[1])
				if err != nil {
					log.Error("invalid port number", "addr", addr, "port", addr[1], "err", err)
					goto loop
				}
			}
			err = cluster.Join(addr[0], port)

			if err != nil {
				log.Error("can't join cluster", "addr", addr, "err", err)
				goto loop
			}
		}
	case txei := <-tch:
		if txe, ok := txei.(*transaction.Envelope); ok {
			log.Debug("received transaction", "txn", txe)
			if err = broadcast(cluster, txe, MSG_TRANSACTION); err != nil {
				log.Error("error broadcasting a transaction message", "txn", txe, "err", err)
			} else {
				log.Debug("broadcasted transaction", "txn", txe)
			}
		}
	case obji := <-och:
		if obj, ok := obji.(git.Object); ok {
			id, err := wendy.NodeIDFromBytes(util.SHA256(obj.Hash()))
			if err != nil {
				log15.Error("error preparing msg id for a git object", "obj", obj, "err", err)
			} else {
				msg := cluster.NewMessage(MSG_REGULAR|MSG_OBJECT, id, git.ObjectToBytes(obj))
				if err = cluster.Send(msg); err != nil {
					log.Error("error sending git object", "obj", obj, "err", err)
				}
			}
		}
	}
	goto loop
}
