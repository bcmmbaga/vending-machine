package api

import (
	"encoding/json"
	"net/http"

	"github.com/bcmmbaga/vending-machine/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

	coll := a.conn.Database(a.config.DatabaseName).Collection("users")
	res := coll.FindOne(c.Request.Context(), bson.M{"username": params.Username})
	if res.Err() != mongo.ErrNoDocuments {
		c.JSON(http.StatusForbidden, gin.H{"message": "username already existed"})
		return
	}

	user, err := models.NewUser(params.Username, params.Password, params.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	_, err = coll.InsertOne(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to save new user"})
		return
	}

	c.JSON(http.StatusCreated, &user)
}
