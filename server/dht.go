package server

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/go-chord"
)

func DHTServer(srv *T) {
	ch := make(chan string)
	_, err := router.PermanentSubscribe("/dht/join", ch)

	transport, err := chord.InitTCPTransport(fmt.Sprintf("0.0.0.0:%d", srv.NetPort), time.Second*10)
	if err != nil {
		log.Printf("Error initializing network transport: %v", err)
		os.Exit(1)
	}

	config := chord.DefaultConfig(srv.NetHostname)
	ring, err := chord.Create(config, transport)
	if err != nil {
		log.Printf("Error initializing the Chord ring: %v", err)
		os.Exit(1)
	}
	log.Println("Ring: %+v", ring)

loop:
	select {
	case existing := <-ch:
		log.Printf("Received a request to join the Chord ring at %s", existing)
		// err = ring.Leave()
		// if err != nil {
		// 	log.Printf("Error leaving the Chord ring: %v", err)
		// }
		// ring.Shutdown()
		ring, err = chord.Join(config, transport, existing)
		if err != nil {
			log.Printf("Error joining existing Chord ring at %s: %v", existing, err)
			ring, err = chord.Create(config, transport)
			if err != nil {
				log.Printf("Error initializing the Chord ring: %v", err)
				os.Exit(1)
			}
		}
		log.Printf("Successfully joined Chord ring at %s", existing)
	}
	goto loop
}
