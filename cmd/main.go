package main

import (
	"log"

	vendingmachine "github.com/bcmmbaga/vending-machine"
	"github.com/bcmmbaga/vending-machine/api"
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
	apiServer := api.NewServer(&serverConfig)

	if err := apiServer.Start(); err != nil {
		log.Fatalln(err.Error())
	}

}
