package api

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/bargez/pktline"
	"github.com/gitchain/gitchain/git"
	"github.com/gitchain/gitchain/repository"
	"github.com/gorilla/mux"
)

func setupGitRoutes(r *mux.Router) {
	// Git Server
	r.Methods("POST").Path("/{path}/git-upload-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(200)
	})

	r.Methods("POST").Path("/{path:.+}/git-receive-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		var lines [][]byte
		dec := pktline.NewDecoder(req.Body)
		dec.DecodeUntilFlush(&lines)

		packfile, err := git.ReadPackfile(req.Body)
		if err != nil {
			log.Printf("error while reading packfile: %v", err)
			resp.WriteHeader(500)
		} else {
			for i := range packfile.Objects {
				err = git.WriteObject(packfile.Objects[i], path.Join(srv.Path, "objects"))
				if err != nil {
					log.Printf("error while writing object: %v", err)
				}
			}
			resp.Header().Add("Cache-Control", "no-cache")
			resp.WriteHeader(200)
		}
	})

	r.Methods("GET").Path("/{path:.+}/info/refs").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		service := req.Form["service"][0]

		path := mux.Vars(req)["path"]
		repo, err := srv.DB.GetRepository(path)
		if err != nil {
			log.Printf("git http protocol: error while retrieving repository %s: %v", path, err)
			resp.WriteHeader(500)
			return
		}
		if repo == nil || repo.Status == repository.PENDING {
			resp.WriteHeader(404)
			return
		}
		resp.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
		resp.Header().Add("Cache-Control", "no-cache")
		enc := pktline.NewEncoder(resp)
		enc.Encode([]byte(fmt.Sprintf("# service=%s\n", service)))
		enc.Encode(nil)
		enc.Encode(append(append(append(append([]byte("0000000000000000000000000000000000000000"), 32), capabilities()...), []byte{0, 32, 10}...)))
		enc.Encode(nil)
	})
}

func capabilities() []byte {
	return []byte("capabilities^{} report-status delete-refs side-band-64k quiet ofs-delta agent=gitchain")
}
