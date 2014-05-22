package api

import (
	"github.com/gitchain/gitchain/server"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/inconshreveable/log15"
)

func jsonRpcService(srv *server.T, log log15.Logger) *rpc.Server {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(&KeyService{srv: srv, log: log}, "")
	s.RegisterService(&NameService{srv: srv, log: log}, "")
	s.RegisterService(&BlockService{srv: srv, log: log}, "")
	s.RegisterService(&TransactionService{srv: srv, log: log}, "")
	s.RegisterService(&RepositoryService{srv: srv, log: log}, "")
	s.RegisterService(&NetService{srv: srv, log: log}, "")
	return s
}
