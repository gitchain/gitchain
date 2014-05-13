package main

import (
	"gitchain/http"
	"gitchain/transaction"
)

func main() {
	transaction.StartListener()
	http.Start()
}
