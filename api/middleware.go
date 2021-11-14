package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// authenticationMiddleware validate content-type of each request is of type application/json
// and Authotization header for all endpoint except user signin
func authenticationMiddleware(c *gin.Context) {
	contType := c.Request.Header.Get("Content-Type")
	if contType != "application/json" {
		c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "unsupported content type",
		})
		return
	}

	fmt.Println(c.Request.RequestURI, c.Request.Method)
	// check for authorization header except for /user URI with POST method.
	if c.Request.RequestURI == "/user" && strings.ToUpper(c.Request.Method) == http.MethodPost {
		c.Next()
	} else {
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader != "" {
			//verify user account and then if account is present then validate
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Missing Authorization header",
			})
			return
		}
	}

}
