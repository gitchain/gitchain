package api

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/bargez/pktline"
	"github.com/gitchain/gitchain/git"
	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/server"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
)

func pktlineToBytes(b []byte) []byte {
	var bytebuf []byte
	buf := bytes.NewBuffer(bytebuf)
	enc := pktline.NewEncoder(buf)
	enc.Encode(b)
	return buf.Bytes()
}
func setupGitRoutes(r *mux.Router, srv *server.T, log log15.Logger) {
	log = log.New("cmp", "git")
	// Git Server
	r.Methods("POST").Path("/{repository:.+}/git-upload-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(200)
	})

	r.Methods("POST").Path("/{repository:.+}/git-receive-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		reponame := mux.Vars(req)["repository"]
		var lines [][]byte
		dec := pktline.NewDecoder(req.Body)
		dec.DecodeUntilFlush(&lines)
		resp.Header().Add("Cache-Control", "no-cache")
		resp.Header().Add("Content-Type", "application/x-git-receive-pack-result")
		enc := pktline.NewEncoder(resp)

		packfile, err := git.ReadPackfile(req.Body)
		if err != nil {
			enc.Encode(append([]byte{1}, pktlineToBytes([]byte(fmt.Sprintf("unpack %v\n", err)))...))
		} else {
			enc.Encode(append([]byte{1}, pktlineToBytes([]byte("unpack ok"))...))
			for i := range packfile.Objects {
				err = git.WriteObject(packfile.Objects[i], path.Join(srv.Config.General.DataPath, "objects"))
				if err != nil {
					enc.Encode(append([]byte{3}, []byte(fmt.Sprintf("Error while writing object: %v\n", err))...))
				} else {
					router.Send("/git/object", make(chan git.Object), packfile.Objects[i])
				}
			}
			for i := range lines {
				split := strings.Split(string(lines[i]), " ")
				old := split[0]
				new := split[1]
				ref := strings.TrimRight(split[2], string([]byte{0}))
				oldHash, err := hex.DecodeString(old)
				if err != nil {
					enc.Encode(append([]byte{3}, []byte(fmt.Sprintf("Malformed hash %s\n", old))...))
					return
				}
				newHash, err := hex.DecodeString(new)
				if err != nil {
					enc.Encode(append([]byte{3}, []byte(fmt.Sprintf("Malformed hash %s\n", new))...))
					return
				}
				tx := transaction.NewReferenceUpdate(reponame, ref, oldHash, newHash)
				key, err := srv.DB.GetMainKey()
				if err != nil {
					enc.Encode(append([]byte{3}, []byte(fmt.Sprintf("Errow while retrieving main key: %v", err))...))
					return
				}
				if key == nil {
					enc.Encode(append([]byte{3}, []byte("No main private key to sign the transaction")...))
					return
				}
				hash, err := srv.DB.GetPreviousEnvelopeHashForPublicKey(&key.PublicKey)
				if err != nil {
					enc.Encode(append([]byte{3}, []byte(fmt.Sprintf("Error while preparing transaction: %v", err))...))
					return
				}

				txe := transaction.NewEnvelope(hash, tx)
				txe.Sign(key)

				enc.Encode(append([]byte{2}, []byte(fmt.Sprintf("[gitchain] Transaction %s\n", txe))...))
				router.Send("/transaction", make(chan *transaction.Envelope), txe)
				enc.Encode(append([]byte{1}, pktlineToBytes([]byte(fmt.Sprintf("ok %s\n", ref)))...))
			}
		}
		enc.Encode(append([]byte{1}, pktlineToBytes(nil)...))
		enc.Encode(nil)

	})

	r.Methods("GET").Path("/{repository:.+}/info/refs").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		service := req.Form["service"][0]

		reponame := mux.Vars(req)["repository"]
		repo, err := srv.DB.GetRepository(reponame)
		if err != nil {
			log.Error("error while retrieving repository", "repo", reponame, "err", err)
			resp.WriteHeader(500)
			return
		}
		if repo == nil || repo.Status == repository.PENDING {
			resp.WriteHeader(404)
			return
		}
		refs, err := srv.DB.ListRefs(reponame)
		if err != nil {
			log.Error("error listing refs", "repo", reponame, "err", err)
			resp.WriteHeader(500)
			return
		}
		reflines := make([][]byte, len(refs))
		for i := range refs {
			ref, err := srv.DB.GetRef(reponame, refs[i])
			if err != nil {
				log.Error("error getting ref", "repo", reponame, "err", err)
				resp.WriteHeader(500)
				return
			}
			refline := append(append([]byte(hex.EncodeToString(ref)), 32), []byte(refs[i])...)
			if i == 0 {
				// append capabilities
				refline = append(append(refline, 0), capabilities()...)
			}
			refline = append(refline, 10) // LF
			reflines[i] = refline
		}

		resp.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
		resp.Header().Add("Cache-Control", "no-cache")
		enc := pktline.NewEncoder(resp)
		enc.Encode([]byte(fmt.Sprintf("# service=%s\n", service)))
		enc.Encode(nil)
		if len(reflines) == 0 {
			enc.Encode(append(append(append(append([]byte("0000000000000000000000000000000000000000"), 32), nulCapabilities()...), []byte{0, 32, 10}...)))
		} else {
			for i := range reflines {
				enc.Encode(reflines[i])
			}
		}
		enc.Encode(nil)
	})
}

func nulCapabilities() []byte {
	return append(append([]byte("capabilities^{}"), 0), capabilities()...)
}

func capabilities() []byte {
	return []byte("report-status delete-refs side-band-64k quiet ofs-delta agent=gitchain")
}
