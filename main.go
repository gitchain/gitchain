package main

import (
	"bytes"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"os"

	"./api"
	"./env"
	"./keys"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "ImportPrivateKey":
		if len(flag.Args()) < 3 {
			fmt.Println("PEM file and alias required: gitchain ImportPrivateKey <file.pem> <alias>")
			os.Exit(1)
		}
		var pemFile = flag.Arg(1)
		var alias = flag.Arg(2)

		content, err := ioutil.ReadFile(pemFile)
		if err != nil {
			fmt.Printf("Can't read %s because of %v", pemFile, err)
			os.Exit(1)
		}
		key, err := keys.ReadPEM(content, true)
		if err != nil {
			fmt.Printf("Can't decode %s because of %v", pemFile, err)
			os.Exit(1)
		}

		http.Post(fmt.Sprintf("http://localhost:%d/key/ImportPrivateKey/%s", env.Port, alias), "application/text",
			bytes.NewReader(pem.EncodeToMemory(key)))
	case "ExportPrivateKey":
		if len(flag.Args()) < 2 {
			fmt.Println("Alias required: gitchain ExportPrivateKey <alias>")
			os.Exit(1)
		}
		var alias = flag.Arg(1)
		resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/key/ExportPrivateKey/%s", env.Port, alias))
		key, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(key))
	case "NameReservation":
		http.Post(fmt.Sprintf("http://localhost:%d/tx/NameReservation", env.Port), "application/json", nil)
	case "Serve":
		StartTransactionListener()
		api.Start()
	default:
		StartTransactionListener()
		api.Start()
	}

}
