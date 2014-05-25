package http

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"path/filepath"

	"github.com/gitchain/gitchain/server/api"
	"github.com/gitchain/gitchain/server/context"
	"github.com/gitchain/gitchain/server/git"
	"github.com/gitchain/gitchain/ui"
	"github.com/gorilla/mux"
)

func Server(srv *context.T) {
	log := srv.Log.New("cmp", "http")

	r := mux.NewRouter()

	// Gitchain API
	r.Methods("POST").Path("/rpc").HandlerFunc(api.JsonRpcService(srv, log).ServeHTTP)
	r.Methods("GET").Path("/info").HandlerFunc(api.InfoHandler(srv, log))

	git.SetupGitRoutes(r, srv, log)

	// UI
	r.Methods("GET").Path("/websocket").HandlerFunc(api.WebsocketHandler(srv, log))
	r.Methods("GET").Path("/").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Content-Type", "text/html")
		var content []byte

		if len(srv.Config.API.DevelopmentModeAssets) > 0 {
			content, _ = ioutil.ReadFile(path.Join(srv.Config.API.DevelopmentModeAssets, "index.html"))
		} else {
			content, _ = ui.Asset("index.html")
		}
		resp.Write(content)
	})
	r.Methods("GET").Path("/{path}").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		file := mux.Vars(req)["path"]
		ext := filepath.Ext(file)
		resp.Header().Add("Content-Type", mime.TypeByExtension(ext))
		var content []byte
		if len(srv.Config.API.DevelopmentModeAssets) > 0 {
			content, _ = ioutil.ReadFile(path.Join(srv.Config.API.DevelopmentModeAssets, file))
		} else {
			content, _ = ui.Asset(file)
		}
		if content == nil {
			resp.WriteHeader(404)
			resp.Write([]byte{})
		} else {
			resp.Write(content)
		}
	})

	http.Handle("/", r)

	err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", srv.Config.API.HttpPort), nil)
	if err != nil {
		log.Crit("error during HTTP server initialization", "err", err)
	}
}
