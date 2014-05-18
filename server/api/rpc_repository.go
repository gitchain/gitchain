package api

import (
	"encoding/hex"
	"net/http"

	"github.com/gitchain/gitchain/repository"
)

type repo struct {
	Name             string
	Status           string
	NameAllocationTx string
}

type RepositoryService struct{}

type ListRepositoriesArgs struct {
}

type ListRepositoriesReply struct {
	Repositories []repo
}

var status = map[int]string{
	repository.PENDING: "pending",
	repository.ACTIVE:  "active",
}

func (*RepositoryService) ListRepositories(r *http.Request, args *ListRepositoriesArgs, reply *ListRepositoriesReply) error {
	repos := srv.DB.ListRepositories()
	for i := range repos {
		r, err := srv.DB.GetRepository(repos[i])
		if err != nil {
			return err
		}
		reply.Repositories = append(reply.Repositories,
			repo{
				Name:             r.Name,
				Status:           status[r.Status],
				NameAllocationTx: hex.EncodeToString(r.NameAllocationTx)})
	}
	return nil
}
