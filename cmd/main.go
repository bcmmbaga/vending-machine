package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bcmmbaga/vending-machine/api"
)

func main() {

	port := ":8080"

	// 	//gracefully shutdown server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		server := api.NewServer()
		if err := server.Start(port); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	<-c
	fmt.Println("Interrupt received gracefully shutting down all services")

}
