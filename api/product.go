package api

import (
	"encoding/json"
	"net/http"

	"github.com/bcmmbaga/vending-machine/models"
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (a *api) NewProduct(c *gin.Context) {
	params := newProductParams{}

	err := c.BindJSON(&params)
	if err != nil {
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			c.JSON(http.StatusBadRequest, syntaxError)
			return
		}
	}

	seller := c.GetString(usernameContext)
	product := models.NewProduct(params.Name, params.Available, params.Cost, seller)

	coll := a.db.Collection("products")
	_, err = coll.InsertOne(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create new Product"})
		return
	}

	c.JSON(http.StatusCreated, product)

}

func (a *api) GetProduct(c *gin.Context) {
	productId, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing productId in URI param"})
		return
	}

	product := models.Product{}

	coll := a.db.Collection("products")
	err := coll.FindOne(c.Request.Context(), bson.M{"_id": productId}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "product not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, product)
}

func (a *api) UpdateProduct(c *gin.Context) {
	productId, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing productId in URI param"})
		return
	}

	params := UpdateProductParams{}

	err := c.BindJSON(&params)
	if err != nil {
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			c.JSON(http.StatusBadRequest, syntaxError)
			return
		}
	}

	product := models.Product{}
	coll := a.db.Collection("products")

	err = coll.FindOne(c.Request.Context(), bson.M{"_id": productId}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "product not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if product.SellerId != c.GetString(usernameContext) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Failed to update product not product owner"})
		return
	}

	updateQuery := bson.M{}
	paramsMap := structs.Map(&params)

	for key := range paramsMap {
		if value := paramsMap[key]; value != "" && value != 0 && value != 0.0 {
			updateQuery[key] = value
		}
	}

	_, err = coll.UpdateOne(c.Request.Context(), bson.M{"_id": productId}, bson.M{"$set": updateQuery})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, nil)

}

func (a *api) DeleteProduct(c *gin.Context) {
	productId, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing productId in URI param"})
		return
	}

	product := models.Product{}
	coll := a.db.Collection("products")

	err := coll.FindOne(c.Request.Context(), bson.M{"_id": productId}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "product not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if product.SellerId != c.GetString(usernameContext) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Failed to delete product not product owner"})
		return
	}

	err = coll.FindOneAndDelete(c.Request.Context(), bson.M{"_id": productId}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Failed to delete product")
		return
	}

	c.JSON(http.StatusOK, product)
}
