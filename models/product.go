package models

import (
	"github.com/google/uuid"
)

type Product struct {
	ID        string `json:"id" bson:"_id"`
	Name      string `json:"name"`
	Available int    `json:"available"`
	Cost      int    `json:"cost"`
	SellerId  string `json:"sellerId"`
}

func NewProduct(name string, available int, cost int, sellerId string) *Product {
	return &Product{
		ID:        uuid.Must(uuid.NewUUID()).String(),
		Name:      name,
		Available: available,
		Cost:      cost,
		SellerId:  sellerId,
	}
}
