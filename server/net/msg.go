package net

import (
	"bytes"
	"encoding/gob"

	"github.com/gitchain/gitchain/types"
	"github.com/gitchain/wendy"
)

const (
	MSG_BROADCAST   byte = 0x80
	MSG_REGULAR     byte = 0x40
	MSG_REPLY       byte = MSG_BROADCAST | MSG_REGULAR
	MSG_TRANSACTION byte = 0x01
	MSG_OBJECT      byte = 0x02
)

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
