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

	"github.com/gitchain/gitchain/server"
	"github.com/gitchain/gitchain/server/api"

	"github.com/gorilla/rpc/json"
)

var configFile, dataPath, assets, netHostname string
var httpPort, netPort int

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
	flag.StringVar(&configFile, "config", "", "path to a config file")
	flag.StringVar(&dataPath, "data-path", "gitchain.db", "path to the data directory, defaults to gitchain.db")
	flag.StringVar(&assets, "development-mode-assets", "", "path to the assets (ui) directory, only for development")
	flag.IntVar(&httpPort, "http-port", 3000, "HTTP port to connect to or serve on")
	flag.IntVar(&netPort, "net-port", 31000, "Gitchain network port to serve on")
	flag.StringVar(&netHostname, "net-hostname", "", "Gitchain network hostname")
	flag.Parse()

	var config *server.Config
	var err error

	config = server.DefaultConfig()

	config.General.DataPath = dataPath
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

	switch flag.Arg(0) {
	case "GeneratePrivateKey":
		if flag.NArg() < 2 {
			fmt.Println("PEM file and alias required: gitchain GeneratePrivateKey <alias>")
			os.Exit(1)
		}
		var alias = flag.Arg(1)

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
	case "SetMainKey":
		if flag.NArg() < 2 {
			fmt.Println("Alias required: gitchain SetMainKey <alias>")
			os.Exit(1)
		}
		var alias = flag.Arg(1)
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

	case "ListPrivateKeys":
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

	case "NameReservation":
		if flag.NArg() < 3 {
			fmt.Println("Command format required: gitchain NameReservation <private key alias> <name>")
			os.Exit(1)
		}
		alias := flag.Arg(1)
		name := flag.Arg(2)
		var resp api.NameReservationReply
		err := jsonrpc(config, "NameService.NameReservation", &api.NameReservationArgs{Alias: alias, Name: name}, &resp)
		if err != nil {
			fmt.Printf("Can't make a name reservation because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Name reservation for %s has been submitted (%s)\nRecord the above transaction hash and following random number for use during allocation: %s\n", name, resp.Id, resp.Random)
	case "NameAllocation":
		if flag.NArg() < 4 {
			fmt.Println("Command format required: gitchain NameReservation <private key alias> <name> <random or reservation tx hash>")
			os.Exit(1)
		}
		alias := flag.Arg(1)
		name := flag.Arg(2)
		random := flag.Arg(3)
		var resp api.NameAllocationReply
		err := jsonrpc(config, "NameService.NameAllocation", &api.NameAllocationArgs{Alias: alias, Name: name, Random: random}, &resp)
		if err != nil {
			fmt.Printf("Can't make a name allocation because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Name allocation for %s has been submitted (%s)\n", name, resp.Id)
	case "ListRepositories":
		var resp api.ListRepositoriesReply
		err := jsonrpc(config, "RepositoryService.ListRepositories", &api.ListRepositoriesArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't list repositories because of %v\n", err)
			os.Exit(1)
		}
		for i := range resp.Repositories {
			fmt.Printf("%s %s %s\n", resp.Repositories[i].Name, resp.Repositories[i].Status, resp.Repositories[i].NameAllocationTx)
		}
	case "LastBlock":
		var resp api.GetLastBlockReply
		err := jsonrpc(config, "BlockService.GetLastBlock", &api.GetLastBlockArgs{}, &resp)
		if err != nil {
			fmt.Printf("Can't get a block because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", resp.Hash)
	case "Block":
		if flag.NArg() < 2 {
			fmt.Println("Command format required: gitchain Block <block hash>")
			os.Exit(1)
		}
		hash := flag.Arg(1)
		var resp api.GetBlockReply
		err := jsonrpc(config, "BlockService.GetBlock", &api.GetBlockArgs{Hash: hash}, &resp)
		if err != nil {
			fmt.Printf("Can't get a block because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Previous block hash: %v\nNext block hash: %v\nMerkle root hash: %v\nTimestamp: %v\nBits: %#x\nNonce: %v\nTransactions: %d\n",
			resp.PreviousBlockHash, resp.NextBlockHash, resp.MerkleRootHash,
			time.Unix(resp.Timestamp, 0).String(), resp.Bits, resp.Nonce, resp.NumTransactions)
	case "Transactions":
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
	case "Transaction":
		if flag.NArg() < 2 {
			fmt.Println("Command format required: gitchain Transaction <transaction hash>")
			os.Exit(1)
		}
		hash := flag.Arg(1)
		var resp api.GetTransactionReply
		err := jsonrpc(config, "TransactionService.GetTransaction", &api.GetTransactionArgs{Hash: hash}, &resp)
		if err != nil {
			fmt.Printf("Can't get a transaction because of %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Previous transaction hash: %v\nPublic key: %v\nNext public key: %v\nValid: %v\n%+v\n",
			resp.PreviousTransactionHash, resp.PublicKey, resp.NextPublicKey, resp.Valid,
			resp.Content)
	case "Info":
		var httpPort int
		flag.IntVar(&httpPort, "http-port", 3000, "HTTP port to connect to or serve on")
		flag.Parse()

		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/info", httpPort))
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
	case "Join":
		if flag.NArg() < 2 {
			fmt.Println("Command format required: gitchain Join <hostname:port>")
			os.Exit(1)
		}
		host := flag.Arg(1)
		var resp api.JoinReply
		err := jsonrpc(config, "NetService.Join", &api.JoinArgs{Host: host}, &resp)
		if err != nil {
			fmt.Printf("Can't join because of %v\n", err)
			os.Exit(1)
		}
	case "Serve":
		fallthrough
	default:
		srv := &server.T{Config: config}
		err := srv.Init()
		if err != nil {
			log.Printf("Error during server initialization: %v", err) // don't use log15 here
			os.Exit(1)
		}
		go server.DHTServer(srv)
		go server.NameRegistrar(srv)
		go server.RepositoryServer(srv)
		go server.MiningFactory(srv)
		go server.TransactionListener(srv)
		api.Start(srv)
	}

}
