package api

import (
	"net/http"
	"strings"

	"github.com/bcmmbaga/vending-machine/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// authenticationMiddleware validate content-type of each request is of type application/json
// and Authotization header for all endpoint except user signin
func (a *api) authenticationMiddleware(c *gin.Context) {
	contType := c.Request.Header.Get("Content-Type")
	if contType != "application/json" {
		c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "unsupported content type",
		})
		return
	}

	// check for authorization header except for /user URI with POST method.
	if (c.Request.RequestURI == "/user" && strings.ToUpper(c.Request.Method) == http.MethodPost) ||
		(c.Request.RequestURI == "/login" && strings.ToUpper(c.Request.Method) == http.MethodPost) {
		c.Next()
	} else {
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader != "" {
			p := jwt.Parser{ValidMethods: []string{jwt.SigningMethodHS256.Name}}
			token, err := p.ParseWithClaims(authHeader, &apiTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(a.config.Secret), nil
			})

			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Invalid authorization token",
				})
				return
			}

			claims, ok := token.Claims.(*apiTokenClaims)
			if !ok && !token.Valid {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Invalid authorization token",
				})
				return
			}

			coll := a.db.Collection("sessions")
			// validate session token if is active
			session := models.Session{}
			err = coll.FindOne(c.Request.Context(), bson.M{"username": claims.Username, "status": "active"}).Decode(&session)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid session token"})
				return
			}

			if session.Token != authHeader && session.Status == "inactive" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid session token"})
				return
			}

			c.Set(usernameContext, claims.Username)
			c.Next()

		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Missing Authorization header",
			})
			return
		}
	}

}

// buyersOnlyMiddleware check whether the user making the request has buyer's role.
func (a *api) buyersOnlyMiddleware(c *gin.Context) {
	username := c.GetString(usernameContext)

	user := models.User{}

	coll := a.db.Collection("users")
	err := coll.FindOne(c.Request.Context(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Account not found"})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to process the request"})
		return
	}

	ok := user.HasRole("buyer")
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Account permission failed"})
		return
	}

	c.Next()
}
