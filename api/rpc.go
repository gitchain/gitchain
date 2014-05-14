package api

import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net/http"

	"../env"
	"../keys"
	"../transaction"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

func jsonRpcService() *rpc.Server {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(KeyService), "")
	s.RegisterService(new(NameService), "")
	return s
}

// KeySerice
type KeyService struct{}

type ImportPrivateKeyArgs struct {
	Alias string
	PEM   []byte
}

type ImportPrivateKeyReply struct {
	Success bool
}

func (srv *KeyService) ImportPrivateKey(r *http.Request, args *ImportPrivateKeyArgs, reply *ImportPrivateKeyReply) error {
	err := env.DB.PutKey(args.Alias, args.PEM)
	if err != nil {
		reply.Success = false
		return err
	}
	reply.Success = true
	return nil
}

type ListPrivateKeysArgs struct {
}

type ListPrivateKeysReply struct {
	Aliases []string
}

func (srv *KeyService) ListPrivateKeys(r *http.Request, args *ListPrivateKeysArgs, reply *ListPrivateKeysReply) error {
	reply.Aliases = env.DB.ListKeys()
	return nil
}

// NameService
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
	pemBlock, err := keys.ReadPEM(env.DB.GetKey(args.Alias), false)
	fmt.Println(pemBlock)
	if err != nil {
		return err
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	fmt.Println(privateKey)
	if err != nil {
		return err
	}
	tx, random := transaction.NewNameReservation(args.Name, &privateKey.PublicKey)
	reply.Id = hex.EncodeToString(tx.Id())
	reply.Random = hex.EncodeToString(random)
	return nil
}
