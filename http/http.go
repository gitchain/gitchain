package http

import (
	"github.com/go-martini/martini"
)

func Start() {

	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	r.Get("/", func() string {
		return "hello world" // HTTP 200 : "hello world"
	})

	m.Run()
}
