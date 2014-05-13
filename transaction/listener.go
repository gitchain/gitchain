package transaction

var ch chan T

func StartListener() {
	ch = make(chan T)
	go listener()
}

func Broadcast(tx T) {
	ch <- tx
}

func listener() {
loop:
	_ = <-ch
	goto loop
}
