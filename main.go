package main

import (
	"bytes"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"os"

	"github.com/gitchain/gitchain/api"
	"github.com/gitchain/gitchain/env"
	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/server"

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
	case "SetMainKey":
		if len(flag.Args()) < 2 {
			fmt.Println("Alias required: gitchain SetMainKey <alias>")
			os.Exit(1)
		}
		var alias = flag.Arg(1)
		var resp api.SetMainKeyReply
		err := jsonrpc("KeyService.SetMainKey", &api.SetMainKeyArgs{Alias: alias}, &resp)
		if err != nil {
			fmt.Printf("Can't set main private key to %s because of %v\n", alias, err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Printf("Can't set main private key to %s (doesn't exist?)\n", alias)
			os.Exit(1)
		}
		fmt.Printf("Successfully set main private key to %s\n", alias)

	case "ListPrivateKeys":
		var resp api.ListPrivateKeysReply
		err := jsonrpc("KeyService.ListPrivateKeys", &api.ListPrivateKeysArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't list private keys because of %v\n", err)
			os.Exit(1)
		}
		var mainKeyResp api.GetMainKeyReply
		err = jsonrpc("KeyService.GetMainKey", &api.GetMainKeyArgs{}, &mainKeyResp)
		if err != nil {
			fmt.Printf("Can't discover main private key because of %v\n", err)
			os.Exit(1)
		}
		for i := range resp.Aliases {
			fmt.Printf("%s %s\n", func() string {
				if resp.Aliases[i] == mainKeyResp.Alias {
					return "*"
				} else {
					return " "
				}
			}(), resp.Aliases[i])
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
	case "LastBlock":
		var resp api.GetLastBlockReply
		err := jsonrpc("BlockService.GetLastBlock", &api.GetLastBlockArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't get a block because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", resp.Hash)
	case "Block":
		if len(flag.Args()) < 2 {
			fmt.Println("Command format required: gitchain Block <block hash>")
			os.Exit(1)
		}
		hash := flag.Arg(1)
		var resp api.GetBlockReply
		err := jsonrpc("BlockService.GetBlock", &api.GetBlockArgs{Hash: hash}, &resp)
		if err != nil {
			fmt.Printf("Can't get a block because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Previous block hash: %v\nMerkle root hash: %v\nTimestamp: %v\nBits: %#x\nNonce: %v\nTransactions: %d\n",
			resp.PreviousBlockHash, resp.MerkleRootHash, resp.Timestamp, resp.Bits, resp.Nonce, resp.NumTransactions)
	case "Info":
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/info", env.Port))
		if err != nil {
			fmt.Printf("Can't retrieve info because of %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Can't retrieve info because of %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(body))
	case "Serve":
		fallthrough
	default:
		go server.StartMiningFactory()
		server.StartTransactionListener()
		api.Start()
	}

}
