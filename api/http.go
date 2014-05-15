package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"../env"
	"../server"

	"github.com/go-martini/martini"
)

type Info struct {
	Mining server.MiningStatus
}

func Start() {

	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	// Gitchain API
	r.Post("/rpc", jsonRpcService().ServeHTTP)
	r.Get("/info", func(resp http.ResponseWriter) string {
		json, err := json.Marshal(Info{Mining: server.GetMiningStatus()})
		if err != nil {
			log.Printf("Error while serving /info: %v", err)
		}
		resp.Header().Add("Content-Type", "application/json")
		return string(json)
	})

	// Git Server
	r.Post("^(?P<path>.*)/git-upload-pack$", func(params martini.Params, req *http.Request) string {
		body, _ := ioutil.ReadAll(req.Body)
		fmt.Println(req, body)
		return params["path"]
	})

	r.Post("^(?P<path>.*)/git-receive-pack$", func(params martini.Params, req *http.Request) string {
		fmt.Println(req)
		return params["path"]
	})

	r.Get("^(?P<path>.*)/info/refs$", func(params martini.Params, req *http.Request) (int, string) {
		body, _ := ioutil.ReadAll(req.Body)
		fmt.Println(req, body)

		return 404, params["path"]
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", env.Port), m))

}
