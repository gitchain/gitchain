package api

import (
	"net/http"
)

type RepositoryService struct{}

type ListRepositoriesArgs struct {
}

type ListRepositoriesReply struct {
	Repositories []string
}

func (*RepositoryService) ListRepositories(r *http.Request, args *ListRepositoriesArgs, reply *ListRepositoriesReply) error {
	reply.Repositories = srv.DB.ListRepositories()
	return nil
}
