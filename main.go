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
	"./server"

	"github.com/gorilla/rpc/json"
)

func jsonrpc(method string, req, res interface{}) error {
	buf, err := json.EncodeClientRequest(method, req)
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/rpc", env.Port), "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	json.DecodeClientResponse(resp.Body, res)
	return nil
}

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
			fmt.Printf("Can't read %s because of %v\n", pemFile, err)
			os.Exit(1)
		}
		key, err := keys.ReadPEM(content, true)
		if err != nil {
			fmt.Printf("Can't decode %s because of %v\n", pemFile, err)
			os.Exit(1)
		}
		var resp api.ImportPrivateKeyReply
		err = jsonrpc("KeyService.ImportPrivateKey", &api.ImportPrivateKeyArgs{Alias: alias, PEM: pem.EncodeToMemory(key)}, &resp)
		if err != nil {
			fmt.Printf("Can't import private key %s because of %v\n", pemFile, err)
			os.Exit(1)
		}
		if resp.Success {
			fmt.Printf("Private key %s has been successfully imported with an alias of %s\n", pemFile, alias)
		} else {
			fmt.Printf("Server can't import private key %s\n", pemFile)
		}

	case "ListPrivateKeys":
		var resp api.ListPrivateKeysReply
		err := jsonrpc("KeyService.ListPrivateKeys", &api.ListPrivateKeysArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't list private keys because of %v\n", err)
			os.Exit(1)
		}
		for i := range resp.Aliases {
			fmt.Println(resp.Aliases[i])
		}

	case "NameReservation":
		if len(flag.Args()) < 3 {
			fmt.Println("Command format required: gitchain NameReservation <private key alias> <name>")
			os.Exit(1)
		}
		alias := flag.Arg(1)
		name := flag.Arg(2)
		var resp api.NameReservationReply
		err := jsonrpc("NameService.NameReservation", &api.NameReservationArgs{Alias: alias, Name: name}, &resp)
		if err != nil {
			fmt.Printf("Can't make a name reservation because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Name reservation for %s has been submitted (%s)\nRecord this random number for use during allocation: %s\n", name, resp.Id, resp.Random)
	case "Serve":
		server.StartTransactionListener()
		api.Start()
	default:
		server.StartTransactionListener()
		api.Start()
	}

}
