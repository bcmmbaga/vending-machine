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

type logInParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *api) logIn(c *gin.Context) {
	params := logInParams{}
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
			c.JSON(http.StatusForbidden, gin.H{"message": "Account username/password is incorrect"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unexpected error occured"})
		return
	}

	authenticated := user.Authenticate(params.Password)

	if !authenticated {
		c.JSON(http.StatusForbidden, gin.H{"message": "Account username/password is incorrect"})
		return
	}

	token, err := newAPIToken(user.Username, a.config.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to initiate session token"})
		return
	}

	session := models.NewSession(params.Username, token)
	coll = a.db.Collection("sessions")

	// check if there is any active session for this user
	res = coll.FindOne(c.Request.Context(), bson.M{"username": params.Username, "status": "active"})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			_, err = coll.InsertOne(c.Request.Context(), session)
			if err != nil {
				c.JSON(http.StatusInsufficientStorage, gin.H{"message": "Failed to save session"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"token": token})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to verify user session",
		})
		return
	}

	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"message": "There is already an active session using your account",
	})

}

func (a *api) revokeAllSessions(c *gin.Context) {
	username := c.GetString(usernameContext)
	coll := a.db.Collection("sessions")

	_, err := coll.UpdateMany(c.Request.Context(), bson.M{"username": username}, bson.M{"$set": bson.M{
		"status": "inactive",
	}})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process logout request"})
		return
	}

	c.JSON(http.StatusOK, nil)
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
