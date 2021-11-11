package vendingmachine

import "context"

type Account interface {
	NewUser(ctx context.Context, username string, password string, role string) (interface{}, error)
	GetUser(ctx context.Context, username string) (interface{}, error)
	UpdateUser(ctx context.Context, username string, metadata interface{}) (interface{}, error)
	DeleteUser(ctx context.Context, username string) error
	Sessions(ctx context.Context, username string) (interface{}, error)
}

type Stock interface {
	NewProduct(ctx context.Context, productName string, amountAvailable int, cost float64) (interface{}, error)
	GetProduct(ctx context.Context, id string) (interface{}, error)
	UpdateProductt(ctx context.Context, id string, metadata interface{}) (interface{}, error)
	DeleteProduct(ctx context.Context, id string) error
}

// Service describe domain service implementation of vending machine.
type Service interface {
	Account
	Stock
}
