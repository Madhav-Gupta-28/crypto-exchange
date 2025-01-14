package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Madhav-Gupta-28/crypto-exchange/server"
)

const ENDPOINT = "http://localhost:3000"

type Client struct {
	http.Client
}

func NewClient() *Client {
	return &Client{
		Client: *http.DefaultClient,
	}
}

type PlaceLimitOrderParams struct {
	UserId int64
	Size   float64
	Price  float64
	Bid    bool
	Market string
	Type   string
}

func (c *Client) PlaceLimitOrder(p PlaceLimitOrderParams) error {
	e := ENDPOINT + "/order"

	params := &server.PlaceOrderRequest{
		UserId: p.UserId,
		Type:   server.OrderType(p.Type), // Limit or Market
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,
		Market: server.Market(p.Market),
	}

	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))

	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	fmt.Printf("%v", resp)
	return nil
}
