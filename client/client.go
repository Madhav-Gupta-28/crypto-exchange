package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Madhav-Gupta-28/crypto-exchange/server"
)

const Endpoint = "http://localhost:3000/"

type Client struct {
	*http.Client
}

func NewClient() *Client {

	return &Client{
		Client: http.DefaultClient,
	}

}

type PlaceOrderRequest struct {
	UserId float64
	Size   float64
	Price  float64
	Bid    bool
	Type   server.OrderType
}

type CancelOrderRequest struct {
	OrderID int64
	Market  server.Market
}

type GetBookRequest struct {
	Market server.Market
}

func (c *Client) ClientPlaceLimitOrder(params *PlaceOrderRequest) (*server.PlaceOrderResponse, error) {

	e := Endpoint + "place-order"

	param := &server.PlaceOrderRequest{
		UserId: params.UserId,
		Type:   params.Type,
		Bid:    params.Bid,
		Size:   params.Size,
		Price:  params.Price,
		Market: server.MarketEth,
	}

	body, err := json.Marshal(param)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	// Add content type header
	req.Header.Set("Content-Type", "application/json")

	// Executing the Requesr
	res, err := c.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	placeorderresponse := &server.PlaceOrderResponse{}

	if err := json.NewDecoder(res.Body).Decode(placeorderresponse); err != nil {
		return nil, err
	}

	return placeorderresponse, nil

}

func (c *Client) ClientCancelOrder(params *CancelOrderRequest) error {

	e := Endpoint + "cancel-order"

	param := &server.CancelOrderRequest{
		OrderID: params.OrderID,
		Market:  params.Market,
	}

	body, err := json.Marshal(param)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, e, bytes.NewReader(body))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Read and parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Printf("Order Status: %s\n", result["msg"])

	return nil

}

func (c *Client) ClientGetBestBid(params *GetBookRequest) (float64, error) {

	e := Endpoint + "book/" + string(params.Market) + "/best-bid"

	req, err := http.NewRequest(http.MethodGet, e, nil)

	if err != nil {
		return 0, err
	}

	res, err := c.Do(req)

	if err != nil {
		return 0, err
	}

	defer res.Body.Close()

	bestBidResponse := &server.PriceResponse{}

	if err := json.NewDecoder(res.Body).Decode(bestBidResponse); err != nil {
		return 0, err
	}

	return bestBidResponse.Price, nil

}

func (c *Client) ClientGetBestAsk(params *GetBookRequest) (float64, error) {

	e := Endpoint + "book/" + string(params.Market) + "/best-ask"

	req, err := http.NewRequest(http.MethodGet, e, nil)

	if err != nil {
		return 0, err
	}

	res, err := c.Do(req)

	if err != nil {
		return 0, err
	}

	defer res.Body.Close()

	bestBidResponse := &server.PriceResponse{}

	if err := json.NewDecoder(res.Body).Decode(bestBidResponse); err != nil {
		return 0, err
	}

	return bestBidResponse.Price, nil

}
