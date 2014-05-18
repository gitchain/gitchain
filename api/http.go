package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gitchain/gitchain/env"
	"github.com/gitchain/gitchain/ui"
	"github.com/gorilla/mux"
)

func Start() {

	r := mux.NewRouter()

	// Gitchain API
	r.Methods("POST").Path("/rpc").HandlerFunc(jsonRpcService().ServeHTTP)
	r.Methods("GET").Path("/info").HandlerFunc(info)

	// Git Server
	r.Methods("POST").Path("/{path}/git-upload-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		body, _ := ioutil.ReadAll(req.Body)
		fmt.Println(req, body)
		resp.Write([]byte(mux.Vars(req)["path"]))
	})

	r.Methods("POST").Path("/{path}/git-receive-pack").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		fmt.Println(req)
		resp.Write([]byte(mux.Vars(req)["path"]))
	})

	r.Methods("GET").Path("/{path}/info/refs").HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		body, _ := ioutil.ReadAll(req.Body)
		fmt.Println(req, body)

		resp.Write([]byte(mux.Vars(req)["path"]))
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

	err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", env.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
