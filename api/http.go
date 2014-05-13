package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"../env"

	"github.com/go-martini/martini"
)

func Start() {

	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	// Gitchain API
	r.Post("/tx/NameReservation", func() string {
		return "wait for it"
	})
	r.Post("/key/ImportPrivateKey/:alias", func(params martini.Params, req *http.Request) (int, string) {
		alias := params["alias"]
		key, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return 500, ""
		} else {
			env.DB.PutKey(alias, key)
			return 200, ""
		}
	})
	r.Get("/key/ExportPrivateKey/:alias", func(params martini.Params, req *http.Request) (int, string) {
		alias := params["alias"]
		key := env.DB.GetKey(alias)
		if key == nil {
			return 404, ""
		} else {
			return 200, string(key)
		}
	})

	// Git Server
	r.Post("^(?P<path>.*)/git-upload-pack$", func(params martini.Params, req *http.Request) string {
		fmt.Println(req)
		return params["path"]
	})

	r.Post("^(?P<path>.*)/git-receive-pack$", func(params martini.Params, req *http.Request) string {
		fmt.Println(req)
		return params["path"]
	})

	r.Get("^(?P<path>.*)/info/refs$", func(params martini.Params, req *http.Request) (int, string) {
		fmt.Println(req)
		return 404, params["path"]
	})

	m.Run()
}
