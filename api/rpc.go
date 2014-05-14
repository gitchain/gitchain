package api

import (
	"net/http"

	"../env"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

func jsonRpcService() *rpc.Server {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(KeyService), "")
	return s
}

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
