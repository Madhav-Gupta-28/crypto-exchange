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

	e.DELETE("/cancel-order", ex.cancelOrder)

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

type CancelOrderRequest struct {
	OrderID int64
	Market  Market
}

func (ex *Exchange) GetOrderByID(market Market, id int64) *orderbook.Order {
	ob := ex.orderbook[market]
	// Search in asks
	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			if order.ID == id {
				return order
			}
		}
	}
	// Search in bids
	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			if order.ID == id {
				return order
			}
		}
	}
	return nil
}

func (ex *Exchange) cancelOrder(c echo.Context) error {

	var cancelOrder CancelOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&cancelOrder); err != nil {
		return err
	}

	ob := ex.orderbook[cancelOrder.Market]

	order := ex.GetOrderByID(cancelOrder.Market, cancelOrder.OrderID)

	if order == nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "order not found"})
	}

	if order.Bid {
		ob.CancelOrder(order)
		return c.JSON(http.StatusOK, map[string]any{"msg": "order cancelled"})
	}

	return nil

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
	ID        int64
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
				ID:        order.ID,
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
				ID:        order.ID,
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
