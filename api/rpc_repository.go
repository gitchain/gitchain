package api

import (
	"net/http"

	"github.com/gitchain/gitchain/env"
)

type RepositoryService struct{}

type ListRepositoriesArgs struct {
}

type ListRepositoriesReply struct {
	Repositories []string
}

func (srv *RepositoryService) ListRepositories(r *http.Request, args *ListRepositoriesArgs, reply *ListRepositoriesReply) error {
	reply.Repositories = env.DB.ListRepositories()
	return nil
}
