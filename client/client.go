package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Madhav-Gupta-28/crypto-exchange/orderbook"
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
	// Market string
	// Type string
}

func (c *Client) PlaceLimitOrder(p *PlaceLimitOrderParams) (*server.PlaceOrderResponse, error) {
	e := ENDPOINT + "/order"

	params := &server.PlaceOrderRequest{
		UserId: p.UserId,
		Type:   server.OrderType("LIMIT"), // Limit or Market
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,
		Market: server.Market("ETH"),
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}

	err = json.NewDecoder(resp.Body).Decode(placeOrderResponse)
	if err != nil {
		return nil, err
	}

	return placeOrderResponse, nil
}

func (c *Client) PlaceMarketOrder(p *PlaceLimitOrderParams) error {
	e := ENDPOINT + "/order"

	params := &server.PlaceOrderRequest{
		UserId: p.UserId,
		Type:   server.OrderType("MARKET"), // Limit or Market
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,
		Market: server.Market("ETH"),
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

	_ = resp
	// fmt.Printf("%v", resp)
	return nil

}

func (c *Client) CancelOrder(orderId int) error {
	e := ENDPOINT + "/order/" + strconv.Itoa(orderId)

	req, err := http.NewRequest(http.MethodDelete, e, nil)
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

func (c *Client) GetBestBidPrice(market server.Market) (float64, error) {
	e := ENDPOINT + "/book/" + string(market) + "/bid"
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	bestBidResponse := &server.BestBidResponse{}
	err = json.NewDecoder(resp.Body).Decode(bestBidResponse)
	if err != nil {
		return 0, err
	}
	return bestBidResponse.Price, nil
}

func (c *Client) GetBestAskPrice(market server.Market) (float64, error) {
	e := ENDPOINT + "/book/" + string(market) + "/ask"
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	bestAskResponse := &server.BestBidResponse{}
	err = json.NewDecoder(resp.Body).Decode(bestAskResponse)
	if err != nil {
		return 0, err
	}
	return bestAskResponse.Price, nil
}

func (c *Client) GetOrdersByUserid(userId int64) ([]*orderbook.Order, error) {
	e := fmt.Sprintf("%s/order/%d", ENDPOINT, userId)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	orders := []*orderbook.Order{}
	err = json.NewDecoder(resp.Body).Decode(&orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (c *Client) GetTrades(market string) ([]*orderbook.Trade, error) {
	e := ENDPOINT + "/trades/" + market
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	trades := []*orderbook.Trade{}
	err = json.NewDecoder(resp.Body).Decode(&trades)
	if err != nil {
		return nil, err
	}
	return trades, nil
}
