package api

import (
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

func jsonRpcService() *rpc.Server {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(KeyService), "")
	s.RegisterService(new(NameService), "")
	s.RegisterService(new(BlockService), "")
	return s
}
