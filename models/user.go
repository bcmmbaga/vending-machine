package models

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"-"`
	Deposit  int    `json:"deposit"`
	Role     string `json:"role"`
}

type Coins []int

var (
	acceptedCoins = map[int]bool{
		5:   true,
		10:  true,
		20:  true,
		50:  true,
		100: true,
	}
)

func NewUser(username string, password string, role string) (*User, error) {

	pwd, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	if role != "buyer" && role != "seller" {
		return nil, errors.New("unknown role type")
	}

	return &User{
		Username: username,
		Password: pwd,
		Role:     role,
		Deposit:  0,
	}, nil

}

func (u *User) AddDeposit(coins Coins) error {
	ok := u.HasRole("buyer")

	if !ok {
		return errors.New("User does not have buyer role")
	}

	for _, coin := range coins {
		if _, ok := acceptedCoins[coin]; !ok {
			return errors.New("Coin not accepted")
		}

		u.Deposit += coin
		fmt.Println(u.Deposit)

	}

	return nil
}

func (u *User) HasRole(roleName string) bool {
	return u.Role == roleName
}

func (u *User) Authenticate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// hashPassword generates a hashed password from a plaintext string
func hashPassword(password string) (string, error) {
	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}
