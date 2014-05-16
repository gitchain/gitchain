package api

import (
	"encoding/hex"
	"net/http"

	"github.com/gitchain/gitchain/env"
	"github.com/gitchain/gitchain/server"
	"github.com/gitchain/gitchain/transaction"
)

type NameService struct{}

type NameReservationArgs struct {
	Alias string
	Name  string
}

type NameReservationReply struct {
	Id     string
	Random string
}

func (srv *NameService) NameReservation(r *http.Request, args *NameReservationArgs, reply *NameReservationReply) error {
	key, err := env.DB.GetKey(args.Alias)
	if err != nil {
		return err
	}
	tx, random := transaction.NewNameReservation(args.Name, &key.PublicKey)
	reply.Id = hex.EncodeToString(tx.Hash())
	reply.Random = hex.EncodeToString(random)
	server.BroadcastTransaction(tx)
	return nil
}
