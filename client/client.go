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

func (c *Client) ClientPlaceLimitOrder(params *PlaceOrderRequest) error {

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
		return err
	}

	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))

	if err != nil {
		return err
	}

	// Add content type header
	req.Header.Set("Content-Type", "application/json")

	// Executing the Requesr
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
