package api

import "github.com/bcmmbaga/vending-machine/models"

type signUpParams struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type depositParams struct {
	Coins models.Coins `json:"coins"`
}

type newProductParams struct {
	Name      string `json:"name" binding:"required"`
	Available int    `json:"available" binding:"required"`
	Cost      int    `json:"cost" binding:"required"`
}

type UpdateProductParams struct {
	Name      string `json:"name,omitempty"`
	Available int    `json:"available,omitempty"`
	Cost      int    `json:"cost,omitempty"`
}
