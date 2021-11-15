package main

import (
	"log"

	vendingmachine "github.com/bcmmbaga/vending-machine"
	"github.com/bcmmbaga/vending-machine/api"
	"github.com/bcmmbaga/vending-machine/storage"
)

var serverConfig vendingmachine.Config

func init() {
	config, err := vendingmachine.LoadConfiguration("")
	if err != nil {
		log.Fatalln(err.Error())
	}

	serverConfig = *config

}

func main() {

	conn, err := storage.Dial(&serverConfig)
	if err != nil {
		log.Fatalln(err.Error())
	}

	apiServer := api.NewServer(&serverConfig, conn)

	if err := apiServer.Start(); err != nil {
		log.Fatalln(err.Error())
	}

}
