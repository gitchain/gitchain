package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/server/context"
	"github.com/gitchain/gitchain/transaction"
	"github.com/inconshreveable/log15"
)

type TransactionService struct {
	srv *context.T
	log log15.Logger
}

type GetTransactionArgs struct {
	Hash string
}

type GetTransactionReply struct {
	PreviousTransactionHash string
	PublicKey               string
	NextPublicKey           string
	Valid                   bool
	Content                 string
}

func (service *TransactionService) GetTransaction(r *http.Request, args *GetTransactionArgs, reply *GetTransactionReply) error {
	hash, err := hex.DecodeString(args.Hash)
	if err != nil {
		return err
	}
	var tx *transaction.Envelope

	tx, err = service.srv.DB.GetTransaction(hash)
	if err != nil {
		return err
	}

	if tx == nil {
		block, err := service.srv.DB.GetTransactionBlock(hash)
		if err != nil {
			return err
		}
		for i := range block.Transactions {
			if block.Transactions[i].Hash().Equals(hash) {
				tx = block.Transactions[i]
				break
			}
		}
	}
	if tx == nil {
		return err
	}
	reply.PreviousTransactionHash = hex.EncodeToString(tx.PreviousEnvelopeHash)
	pk, err := keys.DecodeECDSAPublicKey(tx.PublicKey)
	if err != nil {
		return err
	}
	reply.PublicKey = keys.ECDSAPublicKeyToString(*pk)
	pk, err = keys.DecodeECDSAPublicKey(tx.PublicKey)
	if err != nil {
		return err
	}
	reply.NextPublicKey = keys.ECDSAPublicKeyToString(*pk)
	reply.Valid = tx.Transaction.Valid()
	jsonEncoded, err := json.Marshal(tx.Transaction)
	if err != nil {
		return err
	}
	reply.Content = string(jsonEncoded)
	return nil
}
