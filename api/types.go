package api

type signUpParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
