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
	//var msg T
loop:
	<-ch
	goto loop
}
