package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"os"

	"github.com/alecthomas/kingpin"
	"github.com/gitchain/gitchain/server"
	"github.com/gitchain/gitchain/server/api"
	netserver "github.com/gitchain/gitchain/server/net"

	"github.com/gorilla/rpc/json"
)

func jsonrpc(config *server.Config, method string, req, res interface{}) error {
	buf, err := json.EncodeClientRequest(method, req)
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/rpc", config.API.HttpPort), "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	json.DecodeClientResponse(resp.Body, res)
	return nil
}

func main() {
	var configFile, dataPath, assets, netHostname string
	var httpPort, netPort int

	var alias, repo, random, hash, node string

	app := kingpin.New("gitchain", "Gitchain daemon and command line interface")
	app.Flag("config", "configuration file").Short('c').ExistingFileVar(&configFile)
	app.Flag("data-path", "path to the data directory").Short('d').StringVar(&dataPath)
	app.Flag("development-mode-assets", "path to the assets (ui) directory, only for developmenty").ExistingDirVar(&assets)
	app.Flag("net-hostname", "Gitchain network hostname").StringVar(&netHostname)
	app.Flag("http-port", "HTTTP port to connect to or listen on").IntVar(&httpPort)
	app.Flag("net-port", "Network port to listen").IntVar(&netPort)

	keypairGenerate := app.Command("keypair-generate", "Generates a new keypair")
	keypairGenerate.Arg("alias", "Keypair name to save it under").Required().StringVar(&alias)

	keypairPrimary := app.Command("keypair-primary", "Sets or gets primary keypair")
	keypairPrimary.Arg("alias", "Keypair name to save it under").StringVar(&alias)

	app.Command("keypair-list", "Lists all keypairs")

	nameReservation := app.Command("name-reservation", "Submits a Name Reservation Transaction")
	nameReservation.Arg("alias", "Keypair name to save it under").Required().StringVar(&alias)
	nameReservation.Arg("name", "Repository name to reserve").Required().StringVar(&repo)

	nameAllocation := app.Command("name-allocation", "Submits a Name Allocation Transaction")
	nameAllocation.Arg("alias", "Keypair name to save it under").Required().StringVar(&alias)
	nameAllocation.Arg("name", "Repository name to allocate").Required().StringVar(&repo)
	nameAllocation.Arg("random", "Random number returned by the name-reservation command").Required().StringVar(&random)

	app.Command("repo-list", "Lists all repositories")

	block := app.Command("block", "Renders a block")
	block.Arg("block", "Block hash").Required().StringVar(&hash)

	app.Command("block-last", "Returns last block hash")

	transactions := app.Command("transactions", "Returns a list of transactions in a block")
	transactions.Arg("block", "Block hash").Required().StringVar(&hash)

	transaction := app.Command("transaction", "Renders a transaction")
	transaction.Arg("txn", "Transaction hash").Required().StringVar(&hash)

	app.Command("info", "Returns gitchain node information")

	join := app.Command("node-join", "Connect to another node")
	join.Arg("node", "Node address <host:port>").Required().StringVar(&node)

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	var config *server.Config
	var err error

	config = server.DefaultConfig()

	config.General.DataPath = dataPath
	config.API.HttpPort = httpPort
	config.API.DevelopmentModeAssets = assets
	config.Network.Hostname = netHostname
	config.Network.Port = netPort

	if len(configFile) > 0 {
		err = server.ReadConfig(configFile, config)
		if err != nil {
			log.Printf("Error read config file %s: %v", configFile, err) // don't use log15 here
			os.Exit(1)
		}
	}

	switch command {
	case "keypair-generate":
		var resp api.GeneratePrivateKeyReply
		err := jsonrpc(config, "KeyService.GeneratePrivateKey", &api.GeneratePrivateKeyArgs{Alias: alias}, &resp)
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
	case "keypair-primary":
		if alias != "" {
			var resp api.SetMainKeyReply
			err := jsonrpc(config, "KeyService.SetMainKey", &api.SetMainKeyArgs{Alias: alias}, &resp)
			if err != nil {
				fmt.Printf("Can't set main private key to %s because of %v\n", alias, err)
				os.Exit(1)
			}
			if !resp.Success {
				fmt.Printf("Can't set main private key to %s (doesn't exist?)\n", alias)
				os.Exit(1)
			}
			fmt.Printf("Successfully set main private key to %s\n", alias)
		} else {
			var mainKeyResp api.GetMainKeyReply
			err = jsonrpc(config, "KeyService.GetMainKey", &api.GetMainKeyArgs{}, &mainKeyResp)
			if err != nil {
				fmt.Printf("Can't discover main private key because of %v\n", err)
				os.Exit(1)
			}
			fmt.Println(mainKeyResp.Alias)
		}
	case "keypair-list":
		var resp api.ListPrivateKeysReply
		err := jsonrpc(config, "KeyService.ListPrivateKeys", &api.ListPrivateKeysArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't list private keys because of %v\n", err)
			os.Exit(1)
		}
		var mainKeyResp api.GetMainKeyReply
		err = jsonrpc(config, "KeyService.GetMainKey", &api.GetMainKeyArgs{}, &mainKeyResp)
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
	case "name-reservation":
		var resp api.NameReservationReply
		err := jsonrpc(config, "NameService.NameReservation", &api.NameReservationArgs{Alias: alias, Name: repo}, &resp)
		if err != nil {
			fmt.Printf("Can't make a name reservation because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Name reservation for %s has been submitted (%s)\nRecord the above transaction hash and following random number for use during allocation: %s\n", repo, resp.Id, resp.Random)
	case "name-allocation":
		var resp api.NameAllocationReply
		err := jsonrpc(config, "NameService.NameAllocation", &api.NameAllocationArgs{Alias: alias, Name: repo, Random: random}, &resp)
		if err != nil {
			fmt.Printf("Can't make a name allocation because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Name allocation for %s has been submitted (%s)\n", repo, resp.Id)
	case "repo-list":
		var resp api.ListRepositoriesReply
		err := jsonrpc(config, "RepositoryService.ListRepositories", &api.ListRepositoriesArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't list repositories because of %v\n", err)
			os.Exit(1)
		}
		for i := range resp.Repositories {
			fmt.Printf("%s %s %s\n", resp.Repositories[i].Name, resp.Repositories[i].Status, resp.Repositories[i].NameAllocationTx)
		}
	case "block-last":
		var resp api.GetLastBlockReply
		err := jsonrpc(config, "BlockService.GetLastBlock", &api.GetLastBlockArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't get a block because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", resp.Hash)
	case "block":
		var resp api.GetBlockReply
		err := jsonrpc(config, "BlockService.GetBlock", &api.GetBlockArgs{Hash: hash}, &resp)
		if err != nil {
			fmt.Printf("Can't get a block because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Previous block hash: %v\nNext block hash: %v\nMerkle root hash: %v\nTimestamp: %v\nBits: %#x\nNonce: %v\nTransactions: %d\n",
			resp.PreviousBlockHash, resp.NextBlockHash, resp.MerkleRootHash,
			time.Unix(resp.Timestamp, 0).String(), resp.Bits, resp.Nonce, resp.NumTransactions)
	case "transactions":
		if flag.NArg() < 2 {
			fmt.Println("Command format required: gitchain Transactions <block hash>")
			os.Exit(1)
		}
		hash := flag.Arg(1)
		var resp api.BlockTransactionsReply
		err := jsonrpc(config, "BlockService.BlockTransactions", &api.BlockTransactionsArgs{Hash: hash}, &resp)
		if err != nil {
			fmt.Printf("Can't get a list of block transactions because of %v\n", err)
			os.Exit(1)
		}
		for i := range resp.Transactions {
			fmt.Println(resp.Transactions[i])
		}
	case "transaction":
		var resp api.GetTransactionReply
		err := jsonrpc(config, "TransactionService.GetTransaction", &api.GetTransactionArgs{Hash: hash}, &resp)
		if err != nil {
			fmt.Printf("Can't get a transaction because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Previous transaction hash: %v\nPublic key: %v\nNext public key: %v\nValid: %v\n%+v\n",
			resp.PreviousTransactionHash, resp.PublicKey, resp.NextPublicKey, resp.Valid,
			resp.Content)
	case "info":
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/info", config.API.HttpPort))
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
	case "join":
		var resp api.JoinReply
		err := jsonrpc(config, "NetService.Join", &api.JoinArgs{Host: node}, &resp)
		if err != nil {
			fmt.Printf("Can't join because of %v\n", err)
			os.Exit(1)
		}
	default:
		srv := &server.T{Config: config}
		err := srv.Init()
		if err != nil {
			log.Printf("Error during server initialization: %v", err) // don't use log15 here
			os.Exit(1)
		}
		go netserver.Server(srv)
		go server.NameRegistrar(srv)
		go server.RepositoryServer(srv)
		go server.MiningFactory(srv)
		go server.TransactionListener(srv)
		api.Start(srv)
	}
}
