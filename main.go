package main

import (
	"github.com/gitchain/gitchain/http"
	"github.com/gitchain/gitchain/transaction"
)

func main() {
	transaction.StartListener()
	http.Start()
}
