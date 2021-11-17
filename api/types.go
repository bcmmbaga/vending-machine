package api

import "github.com/bcmmbaga/vending-machine/models"

type signUpParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type depositParams struct {
	Coins models.Coins `json:"coins"`
}
