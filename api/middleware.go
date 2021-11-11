package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func authenticationMiddleware(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")

	if authHeader != "" {
		//verify user account and then if account is present then validate
		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Missing Authorization header",
		})
	}

}
