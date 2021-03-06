package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	vendingmachine "github.com/bcmmbaga/vending-machine"
	"github.com/bcmmbaga/vending-machine/models"
	"github.com/bcmmbaga/vending-machine/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

var userToken = map[string]string{}
var productId string

func TestDeposit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	api, err := setupNewAPIServer()
	assert.NoError(t, err)

	testUsers := api.setupTestCases()

	testCases := []struct {
		username        string
		coins           models.Coins
		responseCode    int
		expectedDeposit int
	}{
		{
			username:     testUsers[0].Username,
			coins:        models.Coins{5, 16, 90},
			responseCode: 400,
		},
		{
			username:     testUsers[1].Username,
			coins:        models.Coins{5, 10, 100},
			responseCode: 403,
		},
		{
			username:        testUsers[0].Username,
			coins:           models.Coins{5, 100, 20},
			responseCode:    200,
			expectedDeposit: 125,
		},
	}

	for _, test := range testCases {
		rr := httptest.NewRecorder()

		buf, err := json.Marshal(&depositParams{
			Coins: test.coins,
		})
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/deposit", bytes.NewBuffer(buf))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", userToken[test.username])
		assert.NoError(t, err)

		api.handler.ServeHTTP(rr, req)

		if rr.Result().StatusCode == http.StatusOK {
			user := models.User{}
			err = json.NewDecoder(rr.Result().Body).Decode(&user)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedDeposit, user.Deposit)

		} else {
			assert.Equal(t, test.responseCode, rr.Result().StatusCode)
		}

	}

	err = api.removeTestCases(testUsers)
	assert.NoError(t, err)

}

func TestBuy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	api, err := setupNewAPIServer()
	assert.NoError(t, err)

	testUsers := api.setupTestCases()

	testCases := []struct {
		productID    string
		username     string
		quantity     int
		responseCode int
		change       []int
	}{
		{
			productID:    productId,
			username:     testUsers[1].Username,
			quantity:     13,
			responseCode: 403,
		},
		{
			productID:    productId,
			username:     testUsers[0].Username,
			quantity:     63,
			responseCode: 400,
		},
		{
			productID:    productId,
			username:     testUsers[0].Username,
			quantity:     15,
			responseCode: 200,
			change:       []int{0, 1, 0, 0, 0},
		},
	}

	for _, test := range testCases {
		rr := httptest.NewRecorder()

		buf, err := json.Marshal(&buyProductParams{
			ProductID: test.productID,
			Quantity:  test.quantity,
		})
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/buy", bytes.NewBuffer(buf))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", userToken[test.username])
		assert.NoError(t, err)

		api.handler.ServeHTTP(rr, req)

		if rr.Result().StatusCode == http.StatusOK {
			resp := buyProductResp{}
			err = json.NewDecoder(rr.Result().Body).Decode(&resp)
			assert.NoError(t, err)

			assert.Equal(t, test.change, resp.Change)

		} else {
			assert.Equal(t, test.responseCode, rr.Result().StatusCode)
		}

	}

	err = api.removeTestCases(testUsers)
	assert.NoError(t, err)
}

func TestGetProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	api, err := setupNewAPIServer()
	assert.NoError(t, err)

	testUsers := api.setupTestCases()

	testCases := []struct {
		productID    string
		username     string
		responseCode int
	}{
		{
			productID:    uuid.Must(uuid.NewUUID()).String(),
			username:     testUsers[0].Username,
			responseCode: 404,
		},
		{
			productID:    productId,
			username:     testUsers[0].Username,
			responseCode: 200,
		},
	}

	for _, test := range testCases {
		rr := httptest.NewRecorder()

		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/product/%s", test.productID), nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", userToken[test.username])
		assert.NoError(t, err)

		api.handler.ServeHTTP(rr, req)

		if rr.Result().StatusCode == http.StatusOK {
			resp := models.Product{}
			err = json.NewDecoder(rr.Result().Body).Decode(&resp)
			assert.NoError(t, err)

			assert.Equal(t, test.productID, resp.ID)

		} else {
			assert.Equal(t, test.responseCode, rr.Result().StatusCode)
		}

	}

	err = api.removeTestCases(testUsers)
	assert.NoError(t, err)
}

func setupNewAPIServer() (*api, error) {
	config, err := vendingmachine.LoadConfiguration("../.env")
	if err != nil {
		return nil, err
	}

	conn, err := storage.Dial(config)
	if err != nil {
		return nil, err
	}

	return NewServer(config, conn), nil

}

func (a *api) setupTestCases() []models.User {
	coll := a.db.Collection("users")

	buyer, _ := models.NewUser("buyer1", "123456", "buyer")
	buyer.AddDeposit(models.Coins{10, 50, 100})

	seller, _ := models.NewUser("seller1", "123456", "seller")

	_, err := coll.InsertMany(context.Background(), []interface{}{
		buyer,
		seller,
	})
	if err != nil {
		log.Fatalf("Failed to setup user test cases: %s", err.Error())
	}

	buyerToken, _ := newAPIToken(buyer.Username, a.config.Secret)
	userToken[buyer.Username] = buyerToken

	sellerToken, _ := newAPIToken(seller.Username, a.config.Secret)
	userToken[seller.Username] = sellerToken

	coll = a.db.Collection("sessions")
	_, err = coll.InsertMany(context.Background(), []interface{}{
		models.NewSession(buyer.Username, buyerToken),
		models.NewSession(seller.Username, sellerToken),
	})
	if err != nil {
		log.Fatalf("Failed to setup session test cases: %s", err.Error())
	}

	coll = a.db.Collection("products")

	product := models.NewProduct("testing", 30, 10, seller.Username)
	productId = product.ID

	_, err = coll.InsertOne(context.Background(), product)
	if err != nil {
		log.Fatalf("Failed to setup product test cases: %s", err.Error())
	}

	return []models.User{*buyer, *seller}

}

func (a *api) removeTestCases(users []models.User) error {
	coll := a.db.Collection("users")

	for _, user := range users {
		_, err := coll.DeleteOne(context.Background(), bson.M{"username": user.Username})
		if err != nil {
			return err
		}
	}

	coll = a.db.Collection("sessions")

	for _, user := range users {
		_, err := coll.DeleteOne(context.Background(), bson.M{"username": user.Username})
		if err != nil {
			return err
		}
	}

	return nil
}
