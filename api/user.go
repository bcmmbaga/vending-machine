package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *api) SignUpNewUser(c *gin.Context) {
	params := signUpParams{}

	err := c.BindJSON(&params)
	if err != nil {
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			c.JSON(http.StatusBadRequest, syntaxError)
			return
		}
	}

	fmt.Println(params)
}
