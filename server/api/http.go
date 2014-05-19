package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/server"
	"github.com/gitchain/gitchain/ui"
	"github.com/gorilla/mux"
)

var srv *server.T

func Start(srvr *server.T) {
	srv = srvr

	r := mux.NewRouter()

	// Gitchain API
	r.Methods("POST").Path("/rpc").HandlerFunc(jsonRpcService().ServeHTTP)
	r.Methods("GET").Path("/info").HandlerFunc(info)

	// Git Server
	r.Methods("POST").Path("/{path}/git-upload-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte(mux.Vars(req)["path"]))
	})

	r.Methods("POST").Path("/{path:.+}/git-receive-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		b, _ := ioutil.ReadAll(req.Body)
		log.Printf("%+v %s", req, b)

		resp.Write([]byte(mux.Vars(req)["path"]))
	})

	r.Methods("GET").Path("/{path:.+}/info/refs").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
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
		b, _ := ioutil.ReadAll(req.Body)
		log.Printf("%+v %s", req, b)
		resp.WriteHeader(200)
		resp.Write([]byte("00860000000000000000000000000000000000000000 capabilities^{} report-status delete-refs side-band-64k quiet ofs-delta agent=git/1.9.3"))
		resp.Write([]byte("0000"))
	})

	// UI
	r.Methods("GET").Path("/websocket").HandlerFunc(websocketHandler)
	r.Methods("GET").Path("/").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Content-Type", "text/html")
		content, _ := ui.Asset("index.html")
		resp.Write(content)
	})
	r.Methods("GET").Path("/{path}").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		path := mux.Vars(req)["path"]
		ext := filepath.Ext(path)
		resp.Header().Add("Content-Type", mime.TypeByExtension(ext))
		content, _ := ui.Asset(path)
		if content == nil {
			resp.WriteHeader(404)
			resp.Write([]byte{})
		} else {
			resp.Write(content)
		}
	})

	http.Handle("/", r)

	err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", srv.HttpPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
