package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type server struct {
	// s vendingmachine.Service
}

// NewServer initiate new http.Handler with API endpoints to serve.
func NewServer() *server {
	return &server{}
}

func makeHandler() http.Handler {
	r := gin.Default()

	r.Use(authenticationMiddleware)

	return r
}

func (s *server) Start(port string) error {
	err := http.ListenAndServe(port, makeHandler())
	if err != nil {
		return err
	}

	return nil
}
