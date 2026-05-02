package rest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const address = "http://localhost:38080"

var client = &http.Client{
	Timeout: 5 * time.Second,
}

var (
	testSneakerID int64
	testAuctionID int
)

type Sneaker struct {
	ID    int    `json:"ID"`
	Brand string `json:"Brand"`
	Model string `json:"Model"`
}

type Auction struct {
	ID           int       `json:"ID"`
	SneakerID    int       `json:"SneakerID"`
	CurrentPrice int64     `json:"CurrentPrice"`
	EndAt        time.Time `json:"EndAt"`
}

type Bid struct {
	ID     int   `json:"ID"`
	UserID int   `json:"UserID"`
	Amount int64 `json:"Amount"`
}

type AuctionDetailsReply struct {
	Auction Auction `json:"auction"`
	Bids    []Bid   `json:"bids"`
}

func TestCatalogAPI(t *testing.T) {
	t.Run("create sneaker success", CatalogCreateSneaker)
	t.Run("create sneaker bad json", CatalogCreateBadJSON)
	t.Run("get sneaker by id", CatalogGetSneakerByID)
	t.Run("get sneaker bad id", CatalogGetBadID)
	t.Run("get all sneakers", CatalogGetAll)
}

func TestAuctionAPI(t *testing.T) {
	require.NotZero(t, testSneakerID, "need sneaker ID to test auctions")

	t.Run("create auction success", AuctionCreateSuccess)
	t.Run("get auction empty bids", AuctionGetDetailsEmpty)
	t.Run("make bid success", AuctionMakeBidSuccess)
	t.Run("make bid too low", AuctionMakeBidTooLow)
	t.Run("get auction with bids", AuctionGetDetailsWithBids)
}

func CatalogCreateSneaker(t *testing.T) {
	reqBody := []byte(`{"brand": "Nike", "model": "Air Jordan 1"}`)
	resp, err := client.Post(address+"/api/sneaker/create", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err, "failed to send request")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "expected OK status")

	var reply struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reply), "decode failed")
	require.Greater(t, reply.ID, int64(0), "ID should be > 0")

	testSneakerID = reply.ID
}

func CatalogCreateBadJSON(t *testing.T) {
	reqBody := []byte(`{"brand": "Nike", "model": `)
	resp, err := client.Post(address+"/api/sneaker/create", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func CatalogGetSneakerByID(t *testing.T) {
	url := fmt.Sprintf("%s/api/sneaker?id=%d", address, testSneakerID)
	resp, err := client.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var sneaker Sneaker
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sneaker))
	require.Equal(t, "Nike", sneaker.Brand)
	require.Equal(t, "Air Jordan 1", sneaker.Model)
}

func CatalogGetBadID(t *testing.T) {
	resp, err := client.Get(address + "/api/sneaker?id=abc")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func CatalogGetAll(t *testing.T) {
	resp, err := client.Get(address + "/api/sneakers")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var sneakers []Sneaker
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sneakers))
	require.GreaterOrEqual(t, len(sneakers), 1, "should have at least one sneaker")
}

func AuctionCreateSuccess(t *testing.T) {
	endAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	reqJSON := fmt.Sprintf(`{"sneaker_id": %d, "start_price": 10000, "end_at": "%s"}`, testSneakerID, endAt)

	resp, err := client.Post(address+"/api/auction/create", "application/json", bytes.NewBuffer([]byte(reqJSON)))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var reply struct {
		AuctionID int `json:"auction_id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reply))
	require.Greater(t, reply.AuctionID, 0)

	testAuctionID = reply.AuctionID
}

func AuctionGetDetailsEmpty(t *testing.T) {
	url := fmt.Sprintf("%s/api/auction?auction_id=%d", address, testAuctionID)
	resp, err := client.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var details AuctionDetailsReply
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&details))

	require.Equal(t, int64(10000), details.Auction.CurrentPrice)
	require.Empty(t, details.Bids, "bids should be empty initially")
}

func AuctionMakeBidSuccess(t *testing.T) {
	reqJSON := fmt.Sprintf(`{"auction_id": %d, "user_id": 99, "amount": 15000}`, testAuctionID)

	resp, err := client.Post(address+"/api/auction/bid", "application/json", bytes.NewBuffer([]byte(reqJSON)))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func AuctionMakeBidTooLow(t *testing.T) {
	reqJSON := fmt.Sprintf(`{"auction_id": %d, "user_id": 100, "amount": 5000}`, testAuctionID)

	resp, err := client.Post(address+"/api/auction/bid", "application/json", bytes.NewBuffer([]byte(reqJSON)))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func AuctionGetDetailsWithBids(t *testing.T) {
	url := fmt.Sprintf("%s/api/auction?auction_id=%d", address, testAuctionID)
	resp, err := client.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var details AuctionDetailsReply
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&details))

	require.Equal(t, int64(15000), details.Auction.CurrentPrice)

	require.Len(t, details.Bids, 1)
	require.Equal(t, int64(15000), details.Bids[0].Amount)
	require.Equal(t, 99, details.Bids[0].UserID)
}
