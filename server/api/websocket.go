package api

import (
	"encoding/json"
	"net/http"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/server"
	"github.com/gorilla/websocket"
	"github.com/inconshreveable/log15"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(srv *server.T, log log15.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.New("cmp", "websocket")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("error upgrading websocket connetion", "err", err)
			return
		}

		ch := make(chan *block.Block)
		_, err = router.PermanentSubscribe("/block", ch)

	loop:
		select {
		case blk := <-ch:
			encoded, err := json.Marshal(blk)
			if err != nil {
				log.Error("error encoding block", "err", err)
				return
			}
			if err = conn.WriteMessage(websocket.TextMessage, encoded); err != nil {
				log.Error("error sending data", "err", err)
				return
			}

		}
		goto loop

	}
}
