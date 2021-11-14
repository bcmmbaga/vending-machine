package storage

import (
	"context"
	"time"

	vendingmachine "github.com/bcmmbaga/vending-machine"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	*mongo.Client
}

// Dial connect to mongo db storage engine usign specified database URI in config.
func Dial(config *vendingmachine.Config) (*Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.DatabaseURI))
	if err != nil {
		return nil, err
	}

	return &Connection{Client: client}, nil
}
