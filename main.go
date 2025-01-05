package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Madhav-Gupta-28/crypto-exchange/orderbook"
	"github.com/labstack/echo/v4"
)

func main() {

	e := echo.New()

	ex := NewExchange()

	e.POST("/order", ex.handlePlaceOrder)
	e.GET("/book/:market", ex.handleGetOrderbook)
	e.DELETE("/order/:orderID", ex.handleCancelOrder)

	e.Start(":3000")

}

type Market string

const (
	MarketETH Market = "ETH"
)

type OrderType string

const (
	LIMITORDER  OrderType = "LIMIT"
	MARKETORDER OrderType = "MARKET"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.Orderbook
}

type OrderResponse struct {
	Id        int64
	Price     float64
	Size      float64
	Bid       bool
	TimeStamp int64
}

type OrderbookData struct {
	TotalBidVolume float64
	TotalAskVolume float64
	Asks           []*OrderResponse
	Bids           []*OrderResponse
}

func NewExchange() *Exchange {
	ex := &Exchange{
		orderbooks: make(map[Market]*orderbook.Orderbook),
	}
	ex.orderbooks[MarketETH] = orderbook.NewOrderbook()
	return ex
}

type PlaceOrderRequest struct {
	Type   OrderType // Limit or Market
	Bid    bool
	Size   float64
	Price  float64
	Market Market
}

type CancelOrderRequest struct {
	OrderId int64
	Market  Market
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placeorderdata PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeorderdata); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	market := Market(placeorderdata.Market)
	ob := ex.orderbooks[market]

	order := orderbook.NewOrder(placeorderdata.Bid, placeorderdata.Size)

	if placeorderdata.Type == LIMITORDER {
		ob.PlaceLimitOrder(placeorderdata.Price, order)
		return c.JSON(http.StatusOK, map[string]string{"message": "Limit Order placed"})

	}

	if placeorderdata.Type == MARKETORDER {
		matches := ob.PlaceMarketOrder(order)
		return c.JSON(http.StatusOK, map[string]string{"matches": fmt.Sprintf("%v", matches)})
	}

	return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Order Type"})

}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {

	var cancelorderdata CancelOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&cancelorderdata); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	ob, ok := ex.orderbooks[MarketETH]

	if !ok {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "Orderbook of This Market not found"})
	}

	ordercanceled := false

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			if order.Id == cancelorderdata.OrderId {
				ob.CancelOrder(order)
				ordercanceled = true
			}

			if ordercanceled {
				return c.JSON(http.StatusOK, map[string]string{"message": "Order canceled"})
			}
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			if order.Id == cancelorderdata.OrderId {
				ob.CancelOrder(order)
				ordercanceled = true
			}

			if ordercanceled {
				return c.JSON(http.StatusOK, map[string]string{"message": "Order canceled"})
			}
		}
	}

	return nil
}

func (ex *Exchange) handleGetOrderbook(c echo.Context) error {

	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "Orderbook of This Market not found"})
	}
	orderbookData := OrderbookData{
		Asks:           []*OrderResponse{},
		Bids:           []*OrderResponse{},
		TotalBidVolume: ob.BidTotalVolumne(),
		TotalAskVolume: ob.AskTotalVolumne(),
	}
	for _, limits := range ob.Asks() {
		for _, orders := range limits.Orders {
			orderresponse := OrderResponse{
				Id:        orders.Id,
				Price:     limits.Price,
				Size:      orders.Size,
				Bid:       orders.Bid,
				TimeStamp: orders.TimeStamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &orderresponse)
		}
	}
	for _, limits := range ob.Bids() {
		for _, orders := range limits.Orders {
			orderresponse := OrderResponse{
				Id:        orders.Id,
				Price:     limits.Price,
				Size:      orders.Size,
				Bid:       orders.Bid,
				TimeStamp: orders.TimeStamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &orderresponse)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)

}
