package main

import (
	"log"

	"github.com/Kevin2cqs/Go-OriginChain/crypto"
	"github.com/Kevin2cqs/Go-OriginChain/network"
)

func main() {
	validatorPrivKey := crypto.GeneratePrivateKey()

	node := createServer("FULL_NODE", &validatorPrivKey, ":3000", []string{"6000"}, ":9000")
	go node.Start()

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

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
}
