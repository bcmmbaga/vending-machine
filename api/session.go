package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bcmmbaga/vending-machine/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type apiTokenClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
}

type signInParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *api) signin(c *gin.Context) {
	params := signInParams{}
	err := c.BindJSON(&params)
	if err != nil {
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			c.JSON(http.StatusBadRequest, syntaxError)
			return
		}
	}

	user := models.User{}
	coll := a.db.Collection("users")

	res := coll.FindOne(c.Request.Context(), bson.M{"username": params.Username})
	err = res.Decode(&user)
	if err != nil {
		if res.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{"message": "account username/password is incorrect"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected error occured"})
		return
	}

	authenticated := user.Authenticate(params.Password)

	if !authenticated {
		c.JSON(http.StatusForbidden, gin.H{"message": "account username/password is incorrect"})
		return
	}

	token, err := newAPIToken(user.Username, a.config.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to initiate session token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// newAPIToken generate API token with 30 days expiring duration for authorizing other request.
func newAPIToken(username string, secret string) (string, error) {
	claims := &apiTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Subject:   "authotization_token",
			Audience:  "vendingmachine",
			ExpiresAt: time.Now().Add(24 * 30 * time.Hour).Unix(),
		},
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
