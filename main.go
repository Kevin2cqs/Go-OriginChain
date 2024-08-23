package main

import (
	"log"

	"github.com/Kevin2cqs/Go-OriginChain/crypto"
	"github.com/Kevin2cqs/Go-OriginChain/network"
)

func main() {
	node := createServer("FULL_NODE", "", ":3000", []string{"6000"}, ":9000")
	go node.start()
	select {} //Block the main thread
}

func createServer(id string, pk *crypto.PrivateKey, addr string, seedNodes []string, apiListenAddr string) *network.Server {
	opts := network.ServerOpts{
		APIListenAddr: apiListenAddr,
		SeedNodes:     seedNodes,
		ListenAddr:    addr,
		PrivateKey:    pk,
		ID:            id,
	}

	s, err := network.newServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
}
