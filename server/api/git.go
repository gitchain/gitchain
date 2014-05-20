package api

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/bargez/pktline"
	"github.com/gitchain/gitchain/git"
	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gorilla/mux"
)

func setupGitRoutes(r *mux.Router) {
	// Git Server
	r.Methods("POST").Path("/{repository:.+}/git-upload-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(200)
	})

	r.Methods("POST").Path("/{repository:.+}/git-receive-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		reponame := mux.Vars(req)["repository"]
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
			for i := range lines {
				split := strings.Split(string(lines[i]), " ")
				old := split[0]
				new := split[1]
				ref := strings.TrimRight(split[2], string([]byte{0}))
				oldHash, err := hex.DecodeString(old)
				if err != nil {
					resp.WriteHeader(400)
					resp.Write([]byte(fmt.Sprintf("Malformed hash %s", old)))
					return
				}
				newHash, err := hex.DecodeString(new)
				if err != nil {
					resp.WriteHeader(400)
					resp.Write([]byte(fmt.Sprintf("Malformed hash %s", new)))
					return
				}
				tx := transaction.NewReferenceUpdate(reponame, ref, oldHash, newHash)
				key, err := srv.DB.GetMainKey()
				if err != nil {
					log.Printf("Errow while retrieving main key: %v", err)
					resp.WriteHeader(500)
					return
				}
				if key == nil {
					log.Printf("No main private key to sign the transaction")
					resp.WriteHeader(500)
					return
				}
				hash, err := srv.DB.GetPreviousEnvelopeHashForPublicKey(&key.PublicKey)
				if err != nil {
					log.Printf("Error while preparing transaction: %v", err)
					resp.WriteHeader(500)
					return
				}

				txe := transaction.NewEnvelope(hash, tx)
				txe.Sign(key)

				router.Send("/transaction", make(chan *transaction.Envelope), txe)
			}
			resp.Header().Add("Cache-Control", "no-cache")
			resp.WriteHeader(200)
		}
	})

	r.Methods("GET").Path("/{repository:.+}/info/refs").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		service := req.Form["service"][0]

		reponame := mux.Vars(req)["repository"]
		repo, err := srv.DB.GetRepository(reponame)
		if err != nil {
			log.Printf("git http protocol: error while retrieving repository %s: %v", repo, err)
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
