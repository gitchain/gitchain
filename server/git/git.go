package git

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/bargez/pktline"
	"github.com/gitchain/gitchain/git"
	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/server/context"
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

func readObject(srv *context.T, h git.Hash) (b []byte, err error) {
	hash := []byte(hex.EncodeToString(h))
	hd := hash[0:2]
	tl := hash[2:]
	b, err = ioutil.ReadFile(path.Join(srv.Config.General.DataPath, "objects", string(hd), string(tl)))
	if err != nil {
		err = fmt.Errorf("object %s is unretrievable: %v", h, err)
		return
	}
	return
}

func processTree(srv *context.T, h git.Hash, haves []git.Hash) (objs []git.Object, err error) {
	b, err := readObject(srv, h)
	if err != nil {
		return
	}
	tree := git.DecodeObject(b).(*git.Tree)
	objs = append(objs, tree)
	for i := range tree.Entries {
		entry := tree.Entries[i]
		var objects []git.Object
		b, err = readObject(srv, entry.Hash)
		if err != nil {
			return
		}
		obj := git.DecodeObject(b)
		switch obj.Type() {
		case "commit":
			objects, err = processCommit(srv, entry.Hash, haves)
		case "tree":
			objects, err = processTree(srv, entry.Hash, haves)
		case "blob":
			objects = []git.Object{obj}
		case "tag":
			objects = []git.Object{obj}
		}
		if err != nil {
			return
		}
		objs = append(objs, objects...)
	}
	return
}
func processCommit(srv *context.T, want git.Hash, haves []git.Hash) (objs []git.Object, err error) {
	for i := range haves {
		if bytes.Compare(want, haves[i]) == 0 {
			return
		}
	}
	b, err := readObject(srv, want)
	if err != nil {
		return
	}
	commit := git.DecodeObject(b).(*git.Commit)
	objs = append(objs, commit)
	tree, err := processTree(srv, commit.Tree, haves)
	if err != nil {
		return
	}
	objs = append(objs, tree...)
	for i := range commit.Parents {
		var objects []git.Object
		objects, err = processCommit(srv, commit.Parents[i], haves)
		if err != nil {
			return
		}
		objs = append(objs, objects...)
	}
	return
}

type pktlineWriter struct {
	encoder *pktline.Encoder
}

func (w *pktlineWriter) Write(p []byte) (n int, err error) {
	err = w.encoder.Encode(p)
	n = len(p)
	return
}

type sideband64Writer struct {
	writer io.Writer
	band   byte
}

func (w *sideband64Writer) Write(p []byte) (n int, err error) {
	var n1 int
	if len(p) > 65519 {
		portions := len(p) / 65519
		for i := 0; i < portions; i++ {
			n1, err = w.writer.Write(append([]byte{w.band}, p[i*65519:65519]...))
			if err != nil {
				return n + n1, err
			}
			n += n1
			n1 = 0
		}
		if len(p)%65119 > 0 {
			n1, err = w.writer.Write(append([]byte{w.band}, p[portions*65519:]...))
			if err != nil {
				return n + n1, err
			}
			n += n1
			n1 = 0
		}
		return n, nil
	} else {
		return w.writer.Write(append([]byte{w.band}, p...))
	}
}

func SetupGitRoutes(r *mux.Router, srv *context.T, log log15.Logger) {
	log = log.New("cmp", "git")
	// Git Server
	r.Methods("POST").Path("/{repository:.+}/git-upload-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		log := log.New("cmp", "git-upload-pack")
		dec := pktline.NewDecoder(req.Body)
		resp.Header().Add("Cache-Control", "no-cache")
		resp.Header().Add("Content-Type", "application/x-git-upload-pack-result")
		enc := pktline.NewEncoder(resp)

		var wants, haves, common []git.Hash
		var objects []git.Object
		wantsRcvd := false

		for {
			var pktline []byte
			if err := dec.Decode(&pktline); err != nil {
				log.Error("error while decoding pkt-line", "err", err)
				return
			}

			switch {
			case pktline == nil:
				switch {
				case !wantsRcvd:
					wantsRcvd = true
				case wantsRcvd:
					for i := range haves {
						_, err := readObject(srv, haves[i])
						if err == nil {
							enc.Encode([]byte(fmt.Sprintf("ACK %x common\n", haves[i])))
							common = append(common, haves[i])
						} else {
							enc.Encode([]byte("NAK\n"))
						}
					}
					haves = make([]git.Hash, 0)
				}
			case bytes.Compare(pktline, []byte("done\n")) == 0:
				if len(common) == 0 {
					enc.Encode([]byte("NAK\n"))
				}
				goto done
			default:
				line := bytes.Split(pktline, []byte{' '})
				h := bytes.TrimSuffix(line[1], []byte{10})
				hash, err := hex.DecodeString(string(h))
				if err != nil {
					enc.Encode(append([]byte{3}, []byte(fmt.Sprintf("error parsing hash %s: %v\n", line[1], err))...))
					return
				}
				if string(line[0]) == "want" {
					wants = append(wants, hash)
				}
				if string(line[0]) == "have" {
					haves = append(haves, hash)
				}
			}
		}
	done:
		var err error
		for i := range wants {
			var objs []git.Object
			objs, err = processCommit(srv, wants[i], common)
			if err != nil {
				enc.Encode(append([]byte{3}, []byte(fmt.Sprintf("%s", err))...))
				return
			}
			objects = append(objects, objs...)
		}
		// filter out duplicates
		seen := make(map[string]bool)
		filteredObjects := make([]git.Object, 0)
		for i := range objects {
			hash := string(objects[i].Hash())
			if !seen[hash] {
				seen[hash] = true
				filteredObjects = append(filteredObjects, objects[i])
			}
		}
		//

		packfile := git.NewPackfile(filteredObjects)
		err = git.WritePackfile(&sideband64Writer{writer: &pktlineWriter{encoder: enc}, band: 1}, packfile)
		if err != nil {
			enc.Encode(append([]byte{3}, []byte(fmt.Sprintf("%s", err))...))
			return
		}

		enc.Encode(append([]byte{1}, pktlineToBytes(nil)...))
		enc.Encode(nil)
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
					srv.Router.Pub(packfile.Objects[i], "/git/object")
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

				enc.Encode(append([]byte{2}, []byte(fmt.Sprintf("[gitchain] Transaction %s\n", txe.Hash()))...))
				srv.Router.Pub(txe, "/transaction")
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

		ref, err := srv.DB.GetRef(reponame, "refs/heads/master")
		if bytes.Compare(ref, make([]byte, 20)) != 0 {
			reflines = append(reflines, append(append(append([]byte(hex.EncodeToString(ref)), 32), []byte("HEAD")...), 10))
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

	r.Methods("GET").Path("/{repository:.+}/HEAD").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		reponame := mux.Vars(req)["repository"]
		ref, err := srv.DB.GetRef(reponame, "refs/heads/master")
		if err != nil {
			log.Error("error while retrieving repository HEAD", "repo", reponame, "err", err)
			resp.WriteHeader(500)
			return
		}
		resp.Header().Add("Content-Type", "text/plain")
		resp.Header().Add("Cache-Control", "no-cache")
		resp.Write([]byte(hex.EncodeToString(ref)))
	})

	r.Methods("GET").Path("/{repository:.+}/objects/{hash:.+}").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
	})

}

func nulCapabilities() []byte {
	return append(append([]byte("capabilities^{}"), 0), capabilities()...)
}

func capabilities() []byte {
	return []byte("report-status delete-refs side-band-64k quiet ofs-delta multi_ack_detailed agent=gitchain")
}
