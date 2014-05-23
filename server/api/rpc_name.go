package api

import (
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/gitchain/gitchain/server"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/util"
	"github.com/inconshreveable/log15"
)

type NameService struct {
	srv *server.T
	log log15.Logger
}

type NameReservationArgs struct {
	Alias string
	Name  string
}

type NameReservationReply struct {
	Id     string
	Random string
}

func (service *NameService) NameReservation(r *http.Request, args *NameReservationArgs, reply *NameReservationReply) error {
	log := service.log.New("cmp", "api_name")
	key, err := service.srv.DB.GetKey(args.Alias)
	if err != nil {
		return err
	}
	if key == nil {
		return errors.New("can't find the key")
	}
	tx, random := transaction.NewNameReservation(args.Name)

	hash, err := service.srv.DB.GetPreviousEnvelopeHashForPublicKey(&key.PublicKey)
	if err != nil {
		log.Error("error while preparing transaction", "err", err)
	}
	txe := transaction.NewEnvelope(hash, tx)
	txe.Sign(key)

	reply.Id = hex.EncodeToString(txe.Hash())
	reply.Random = hex.EncodeToString(random)
	// We save sha(random+name)=txhash to scraps to be able to find
	// the transaction hash by random and number during allocation
	service.srv.DB.PutScrap(util.SHA256(append(random, []byte(args.Name)...)), txe.Hash())
	service.srv.Router.Pub(txe, "/transaction")
	return nil
}

type NameAllocationArgs struct {
	Alias  string
	Name   string
	Random string
}

type NameAllocationReply struct {
	Id string
}

func (service *NameService) NameAllocation(r *http.Request, args *NameAllocationArgs, reply *NameAllocationReply) error {
	log := service.log.New("cmp", "api_name")
	key, err := service.srv.DB.GetKey(args.Alias)
	if err != nil {
		return err
	}
	if key == nil {
		return errors.New("can't find the key")
	}
	random, err := hex.DecodeString(args.Random)
	if err != nil {
		return err
	}
	tx, err := transaction.NewNameAllocation(args.Name, random)
	if err != nil {
		return err
	}

	hash, err := service.srv.DB.GetPreviousEnvelopeHashForPublicKey(&key.PublicKey)
	if err != nil {
		log.Error("error while preparing transaction", "err", err)
	}

	txe := transaction.NewEnvelope(hash, tx)
	txe.Sign(key)

	reply.Id = hex.EncodeToString(txe.Hash())
	service.srv.Router.Pub(txe, "/transaction")
	return nil
}
