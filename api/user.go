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

	coll := a.db.Collection("users")
	res := coll.FindOne(c.Request.Context(), bson.M{"username": params.Username})
	if res.Err() != mongo.ErrNoDocuments {
		c.JSON(http.StatusForbidden, gin.H{"message": "Username already existed"})
		return
	}

	user, err := models.NewUser(params.Username, params.Password, params.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	_, err = coll.InsertOne(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save new user"})
		return
	}

	c.JSON(http.StatusCreated, &user)
}

func (a *api) GetUser(c *gin.Context) {
	username := c.GetString(usernameContext)

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing username in request"})
		return
	}

	user := models.User{}

	coll := a.db.Collection("users")
	err := coll.FindOne(c.Request.Context(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func (a *api) ResetDeposit(c *gin.Context) {
	coll := a.db.Collection("users")

	username := c.GetString(usernameContext)

	_, err := coll.UpdateOne(c.Request.Context(), bson.M{"username": username}, bson.M{"$set": bson.M{"deposit": 0}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to reset deposit"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (a *api) DeleteUser(c *gin.Context) {
	username := c.GetString(usernameContext)

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing username in request uri"})
		return
	}

	user := models.User{}

	coll := a.db.Collection("users")
	err := coll.FindOneAndDelete(c.Request.Context(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete user acccount"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (a *api) deposit(c *gin.Context) {
	params := depositParams{}

	err := c.BindJSON(&params)
	if err != nil {
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			c.JSON(http.StatusBadRequest, syntaxError)
			return
		}
	}

	user := models.User{}
	username := c.GetString(usernameContext)

	coll := a.db.Collection("users")
	err = coll.FindOne(c.Request.Context(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = user.AddDeposit(params.Coins)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	_, err = coll.UpdateOne(c.Request.Context(), bson.M{"username": username}, bson.M{"$set": bson.M{"deposit": user.Deposit}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to deposit"})
		return
	}

	c.JSON(http.StatusOK, user)
}
