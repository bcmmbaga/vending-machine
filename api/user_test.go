package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	vendingmachine "github.com/bcmmbaga/vending-machine"
	"github.com/bcmmbaga/vending-machine/models"
	"github.com/bcmmbaga/vending-machine/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

var userToken = map[string]string{}

func TestDeposit(t *testing.T) {
	// gin.SetMode(gin.TestMode)

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
			// assert.Equal(t, "hdsfdsf", rr.Body.String())
			err = json.NewDecoder(rr.Result().Body).Decode(&user)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedDeposit, user.Deposit)

		} else {
			assert.Equal(t, test.responseCode, rr.Result().StatusCode)
			t.Log()
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
	seller, _ := models.NewUser("seller1", "123456", "seller")

	_, err := coll.InsertMany(context.Background(), []interface{}{
		buyer,
		seller,
	})
	if err != nil {
		log.Fatalf("Failed to setup test cases: %s", err.Error())
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
		log.Fatalf("Failed to setup test cases: %s", err.Error())
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
