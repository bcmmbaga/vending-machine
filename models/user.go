package models

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"-"`
	Deposit  int    `json:"deposit"`
	Role     string `json:"role"`
}

func NewUser(username string, password string, role string) (*User, error) {

	pwd, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		Username: username,
		Password: pwd,
		Role:     password,
		Deposit:  0,
	}, nil

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
