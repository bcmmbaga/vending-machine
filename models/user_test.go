package models

import "testing"

func TestPasswordAuthentication(t *testing.T) {
	hashPwd, err := hashPassword("testing")
	if err != nil {
		t.Errorf("failed to hash password: %s", err.Error())
	}

	user := User{
		Username: "bcmmbaga",
		Password: hashPwd,
		Deposit:  0,
		Role:     "",
	}

	authenticated := user.Authenticate("testing")
	if !authenticated {
		t.Errorf("failed to authenticate user: %s", err.Error())
	}
}
