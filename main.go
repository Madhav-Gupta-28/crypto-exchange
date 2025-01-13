package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Madhav-Gupta-28/crypto-exchange/orderbook"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)

// All Constanats Defined here
const (
	exchangePrivateKey           = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
	exchangeAddress              = "0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1"
	LIMITORDER         OrderType = "LIMIT"
	MARKETORDER        OrderType = "MARKET"
	MarketETH          Market    = "ETH"
)

// All Type Defined here
type (
	OrderType string
	Market    string

	Exchange struct {
		client     *ethclient.Client
		Users      map[int64]*User
		orders     map[int64]int64
		orderbooks map[Market]*orderbook.Orderbook
		PrivateKey *ecdsa.PrivateKey
	}

	OrderResponse struct {
		Id        int64
		Price     float64
		Size      float64
		Bid       bool
		TimeStamp int64
	}

	OrderbookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Asks           []*OrderResponse
		Bids           []*OrderResponse
	}

	PlaceOrderRequest struct {
		UserId int64
		Type   OrderType // Limit or Market
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	CancelOrderRequest struct {
		OrderId int64
		Market  Market
	}

	MatchedOrder struct {
		Price float64
		Size  float64
		Id    int64
	}

	User struct {
		Id         int64
		PrivateKey *ecdsa.PrivateKey
	}
)

func main() {

	e := echo.New()

	e.HTTPErrorHandler = httpErrorHandler

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex := NewExchange("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d ", client)

	pv1, err := crypto.HexToECDSA("6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1")
	if err != nil {
		log.Fatal(err)
	}

	user1 := &User{
		Id:         69,
		PrivateKey: pv1,
	}
	ex.Users[user1.Id] = user1

	e.POST("/order", ex.handlePlaceOrder)
	e.GET("/book/:market", ex.handleGetOrderbook)
	e.DELETE("/order/:orderID", ex.handleCancelOrder)

	// // Getting rhe balance of the account
	// account := common.HexToAddress("0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1")
	// balance, err := client.BalanceAt(context.Background(), account, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(balance)

	e.Start(":3000")

}

func httpErrorHandler(err error, c echo.Context) {

	fmt.Println(err)

}

func NewExchange(privateKey string, client *ethclient.Client) *Exchange {
	pv, err := crypto.HexToECDSA(exchangePrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	ex := &Exchange{
		client:     client,
		Users:      make(map[int64]*User),
		orders:     make(map[int64]int64),
		orderbooks: make(map[Market]*orderbook.Orderbook),
		PrivateKey: pv,
	}
	ex.orderbooks[MarketETH] = orderbook.NewOrderbook()
	return ex
}

func NewUser(privateKey string) *User {
	pv, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatal(err)
	}
	return &User{
		PrivateKey: pv,
	}
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) []*MatchedOrder {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchedOrders := make([]*MatchedOrder, len(matches))

	isBid := false
	if order.Bid {
		isBid = true
	}
	for i := 0; i < len(matches); i++ {
		id := matches[i].Bid.Id
		if !isBid {
			id = matches[i].Ask.Id
		}
		matchedOrders[i] = &MatchedOrder{
			Price: matches[i].Price,
			Size:  matches[i].SizeFilled,
			Id:    id,
		}
	}

	return matchedOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	user := ex.Users[order.UserId]

	// Transfer the ETH from the user to the exchange
	err2 := TransferETH(ex.client, user.PrivateKey, exchangeAddress, order.Size)
	if err2 != nil {
		return err2
	}
	return nil
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placeorderdata PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeorderdata); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	market := Market(placeorderdata.Market)
	order := orderbook.NewOrder(placeorderdata.Bid, placeorderdata.Size, placeorderdata.UserId)

	if placeorderdata.Type == LIMITORDER {
		err := ex.handlePlaceLimitOrder(market, placeorderdata.Price, order)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Limit Order placed"})
	}
	if placeorderdata.Type == MARKETORDER {
		matches := ex.handlePlaceMarketOrder(market, order)
		return c.JSON(http.StatusOK, map[string]any{"message": "Market Order placed", "matchedOrders": matches})
	}

	return nil
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {

	idstr := c.Param("orderID")
	id, _ := strconv.Atoi(idstr)

	ob := ex.orderbooks[MarketETH]
	order := ob.Orders[int64(id)]
	ob.CancelOrder(order)
	return c.JSON(http.StatusOK, map[string]string{"message": "Order canceled"})
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
