package models

import (
	"errors"
	"sort"

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

func (p *Product) Change(quatity int, amount int) ([]int, error) {

	totalSpent := p.Cost * quatity

	if totalSpent > amount {
		return nil, errors.New("Insufficient balance to complete the purchase")
	}

	changeCoins := makeChange(amount - totalSpent)

	p.Available = p.Available - quatity

	return changeCoins, nil
}

func makeChange(change int) []int {
	if change == 0 {
		return nil
	}

	coinsPool := make([]int, len(acceptedCoins))

	copy(coinsPool, acceptedCoins)
	sort.Sort(sort.Reverse(sort.IntSlice(coinsPool)))

	changeCoinCount := make([]int, len(coinsPool))
	for i, coin := range coinsPool {
		changeCoinCount[i] = change / coin
		change = change - changeCoinCount[i]*coin
	}

	for i, j := 0, len(changeCoinCount)-1; i < j; i, j = i+1, j-1 {
		changeCoinCount[i], changeCoinCount[j] = changeCoinCount[j], changeCoinCount[i]
	}

	return changeCoinCount
}
