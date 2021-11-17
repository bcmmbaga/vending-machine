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
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	usernameContext = "username"
)

type api struct {
	// s vendingmachine.Service
	handler http.Handler

	db *mongo.Database

	config *vendingmachine.Config
}

// NewServer initiate new http.Handler with API endpoints to serve.
func NewServer(config *vendingmachine.Config, conn *storage.Connection) *api {
	api := &api{config: config, db: conn.Database(config.DatabaseName)}

	r := gin.Default()

	r.Use(api.authenticationMiddleware)

	user := r.Group("/user")
	user.GET("", api.GetUser)
	user.POST("", api.SignUpNewUser)
	user.DELETE("", api.DeleteUser)

	r.POST("/deposit", api.buyersOnlyMiddleware(), api.deposit)
	r.POST("/login", api.logIn)
	r.POST("/logout", api.revokeAllSessions)
	r.POST("/logout/all", api.revokeAllSessions)
	r.POST("/reset", api.buyersOnlyMiddleware(), api.ResetDeposit)

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

		s.db.Client().Disconnect(ctx)
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
