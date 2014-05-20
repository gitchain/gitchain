package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
)

type TransactionService struct{}

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

func (*TransactionService) GetTransaction(r *http.Request, args *GetTransactionArgs, reply *GetTransactionReply) error {
	hash, err := hex.DecodeString(args.Hash)
	if err != nil {
		return err
	}
	var tx *transaction.Envelope

	tx, err = srv.DB.GetTransaction(hash)
	if err != nil {
		return err
	}

	if tx == nil {
		block, err := srv.DB.GetTransactionBlock(hash)
		if err != nil {
			return err
		}
		for i := range block.Transactions {
			if types.HashEqual(block.Transactions[i].Hash(), hash) {
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
