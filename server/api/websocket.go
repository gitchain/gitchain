package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/router"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	ch := make(chan *block.Block)
	_, err = router.PermanentSubscribe("/block", ch)

loop:
	select {
	case blk := <-ch:
		encoded, err := json.Marshal(blk)
		if err != nil {
			log.Println(err)
			return
		}
		if err = conn.WriteMessage(websocket.TextMessage, encoded); err != nil {
			log.Println(err)
			return
		}

	}
	goto loop

}
