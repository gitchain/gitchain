package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"os"

	"github.com/gitchain/gitchain/api"
	"github.com/gitchain/gitchain/env"
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
	case "GeneratePrivateKey":
		if len(flag.Args()) < 2 {
			fmt.Println("PEM file and alias required: gitchain GeneratePrivateKey <alias>")
			os.Exit(1)
		}
		var alias = flag.Arg(1)

		var resp api.GeneratePrivateKeyReply
		err := jsonrpc("KeyService.GeneratePrivateKey", &api.GeneratePrivateKeyArgs{Alias: alias}, &resp)
		if err != nil {
			fmt.Printf("Can't generate private key because of %v\n", err)
			os.Exit(1)
		}
		if resp.Success {
			fmt.Printf("Private key has been successfully generated with an alias of %s, the public address is %s\n", alias, resp.PublicKey)
		} else {
			fmt.Printf("Server can't generate the private key\n")
			os.Exit(1)
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
		fmt.Printf("Name reservation for %s has been submitted (%s)\nRecord the above transaction hash and following random number for use during allocation: %s\n", name, resp.Id, resp.Random)
	case "NameAllocation":
		if len(flag.Args()) < 4 {
			fmt.Println("Command format required: gitchain NameReservation <private key alias> <name> <random or reservation tx hash>")
			os.Exit(1)
		}
		alias := flag.Arg(1)
		name := flag.Arg(2)
		random := flag.Arg(3)
		var resp api.NameAllocationReply
		err := jsonrpc("NameService.NameAllocation", &api.NameAllocationArgs{Alias: alias, Name: name, Random: random}, &resp)
		if err != nil {
			fmt.Printf("Can't make a name allocation because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Name allocation for %s has been submitted (%s)\n", name, resp.Id)
	case "ListRepositories":
		var resp api.ListRepositoriesReply
		err := jsonrpc("RepositoryService.ListRepositories", &api.ListRepositoriesArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't list repositories because of %v\n", err)
			os.Exit(1)
		}
		for i := range resp.Repositories {
			fmt.Printf("%s\n", resp.Repositories[i])
		}
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
		fmt.Printf("Previous block hash: %v\nNext block hash: %v\nMerkle root hash: %v\nTimestamp: %v\nBits: %#x\nNonce: %v\nTransactions: %d\n",
			resp.PreviousBlockHash, resp.NextBlockHash, resp.MerkleRootHash,
			time.Unix(resp.Timestamp, 0).String(), resp.Bits, resp.Nonce, resp.NumTransactions)
	case "Transactions":
		if len(flag.Args()) < 2 {
			fmt.Println("Command format required: gitchain Transactions <block hash>")
			os.Exit(1)
		}
		hash := flag.Arg(1)
		var resp api.BlockTransactionsReply
		err := jsonrpc("BlockService.BlockTransactions", &api.BlockTransactionsArgs{Hash: hash}, &resp)
		if err != nil {
			fmt.Printf("Can't get a list of block transactions because of %v\n", err)
			os.Exit(1)
		}
		for i := range resp.Transactions {
			fmt.Println(resp.Transactions[i])
		}
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
		go server.MiningFactory()
		go server.NameRegistrar()
		go server.TransactionListener()
		api.Start()
	}

}
