package api

import (
	"encoding/json"
	"fmt"
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

	params := updateProductParams{}

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

func (a *api) buyProduct(c *gin.Context) {
	params := buyProductParams{}

	err := c.BindJSON(&params)
	if err != nil {
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			c.JSON(http.StatusBadRequest, syntaxError)
			return
		}
	}

	productColl := a.db.Collection("products")

	product := models.Product{}
	err = productColl.FindOne(c.Request.Context(), bson.M{"_id": params.ProductID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "product not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// check if products is available depending on buyer quantity
	if params.Quantity > product.Available {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Product quantity left is %d", product.Available)})
		return
	}

	buyer := models.User{}
	username := c.GetString(usernameContext)
	usersColl := a.db.Collection("users")

	err = usersColl.FindOne(c.Request.Context(), bson.M{"username": username}).Decode(&buyer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Buyer not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	seller := models.User{}
	err = usersColl.FindOne(c.Request.Context(), bson.M{"username": product.SellerId}).Decode(&seller)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Seller not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	totalAmount := params.Quantity * product.Cost
	if totalAmount > buyer.Deposit {
		c.JSON(http.StatusForbidden, gin.H{"message": "Deposit balance is not enough, please make deposit to complete the purchase"})
		return
	}

	// deduce totalAMount from buyer's deposit wallet
	_, err = usersColl.UpdateOne(c.Request.Context(), bson.M{"username": buyer.Username}, bson.M{"$set": bson.M{"deposit": buyer.Deposit - totalAmount}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process the request"})
		return
	}

	// credit the amount to seller deposit account
	_, err = usersColl.UpdateOne(c.Request.Context(), bson.M{"username": seller.Username}, bson.M{"$set": bson.M{"deposit": seller.Deposit + totalAmount}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process the request"})
		return
	}

	// update available product quantity after purchase
	_, err = productColl.UpdateOne(c.Request.Context(), bson.M{"_id": params.ProductID}, bson.M{"$set": bson.M{"available": product.Available - params.Quantity}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process the request"})
		return
	}

	change, _ := product.Change(params.Quantity, buyer.Deposit)

	c.JSON(http.StatusOK, &buyProductResp{
		TotalSpent:      totalAmount,
		ProductName:     product.Name,
		ProductQuantity: params.Quantity,
		Change:          change,
	})
}
