package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	vendingmachine "github.com/bcmmbaga/vending-machine"
	"github.com/bcmmbaga/vending-machine/storage"
	"github.com/gin-gonic/gin"
)

type api struct {
	// s vendingmachine.Service
	handler http.Handler

	client *storage.Connection

	config *vendingmachine.Config
}

// NewServer initiate new http.Handler with API endpoints to serve.
func NewServer(config *vendingmachine.Config) *api {
	r := gin.Default()

	r.Use(authenticationMiddleware)

	api := &api{config: config}

	user := r.Group("/users")
	user.POST("", api.SignUpNewUser)

	api.handler = r

	return api
}

func (s *api) Start() error {
	server := &http.Server{
		Addr:    s.config.Port,
		Handler: s.handler,
	}

	c := make(chan os.Signal, 1)
	go func() {
		waitForTermination(c)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		s.client.Disconnect(ctx)

		ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		server.Shutdown(ctx)
	}()

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func waitForTermination(done <-chan os.Signal) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-signals:
		log.Fatalf("Triggering shutdown from signal %s", sig)
	case <-done:
		log.Println("Shutting down...")
	}
}
