package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Madhav-Gupta-28/crypto-exchange/orderbook"
	"github.com/labstack/echo/v4"
)

func main() {
	fmt.Println("Hello world")

	// Echo Instance
	e := echo.New()
	ex := NewExchange()

	e.GET("/book", ex.handleGetBook)

	e.POST("/place-order", ex.handlePlaceOrder)

	e.Start(":3000")

}

type Market string

const (
	MarketEth Market = "ETH"
)

type Exchange struct {
	orderbook map[Market]*orderbook.Orderbook
}

func NewExchange() *Exchange {

	orderbooks := make(map[Market]*orderbook.Orderbook)

	orderbooks[MarketEth] = orderbook.NewOrderbook()

	return &Exchange{
		orderbook: orderbooks,
	}
}

type OrderType string

const (
	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
)

type PlaceOrderRequest struct {
	Type   OrderType // limit or market
	Bid    bool
	Size   float64
	Price  float64
	Market Market
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placemarkerorder PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placemarkerorder); err != nil {
		return err
	}

	ob := ex.orderbook[Market(placemarkerorder.Market)]

	order := orderbook.NewOrder(placemarkerorder.Bid, placemarkerorder.Size)

	if placemarkerorder.Type == OrderType(LimitOrder) {
		ob.PlaceLimitOrder(placemarkerorder.Price, order)
		return c.JSON(200, map[string]any{"msg": " limit  order placed"})

	}

	if placemarkerorder.Type == OrderType(MarketOrder) {
		matches := ob.PlaceMarketOrder(order)
		return c.JSON(200, map[string]any{"mathces": len(matches)})
	}

	return nil

}

type Order struct {
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type OrderbookData struct {
	TotalBidVolumne float64
	TotalAskVolumne float64
	Asks            []*Order
	Bids            []*Order
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := c.QueryParam("market")

	ob, ok := ex.orderbook[Market(market)]

	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "market not found"})
	}

	orderbookData := OrderbookData{
		TotalBidVolumne: ob.BidTotalVolumne(),
		TotalAskVolumne: ob.AskTotalVolumne(),
		Asks:            []*Order{},
		Bids:            []*Order{},
	}

	for _, limit := range ob.Asks() {

		for _, order := range limit.Orders {
			o := &Order{
				Price:     limit.Price,
				Size:      order.Size(),
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, o)
			// orderbookData.Bids = append(orderbookData.Bids, o)
		}
	}

	for _, limit := range ob.Bids() {

		for _, order := range limit.Orders {
			o := &Order{
				Price:     limit.Price,
				Size:      order.Size(),
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, o)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)

}
