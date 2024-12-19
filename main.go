package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Madhav-Gupta-28/crypto-exchange/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	// "github.com/Madhav-Gupta-28/crypto-exchange/util" // Adjust the path as necessary
)

func main() {
	fmt.Println("Hello world")

	// Echo Instance
	e := echo.New()

	e.HTTPErrorHandler = httpErrorHandler

	client, _ := ethclient.Dial("http://localhost:8545")

	ex := NewExchange("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d", client)

	privateKey, err := crypto.HexToECDSA("6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1")

	if err != nil {
		log.Fatal(err)
	}

	user := &User{
		Id:         9,
		PrivateKey: privateKey,
	}

	ex.Users[user.Id] = user

	fmt.Println(ex.Users)

	address := "0xFFcf8FDEE72ac11b5c542428B35EEF5769C409f0"

	balance, _ := ex.client.BalanceAt(context.Background(), common.HexToAddress(address), nil)

	fmt.Println(balance)

	e.GET("/book", ex.handleGetBook)

	e.DELETE("/cancel-order", ex.cancelOrder)

	e.POST("/place-order", ex.handlePlaceOrder)

	e.Start(":3000")

}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

// Type Decleratio
type (
	Market    string
	OrderType string

	CancelOrderRequest struct {
		OrderID int64
		Market  Market
	}

	PlaceOrderRequest struct {
		UserId float64
		Type   OrderType // limit or market
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	Exchange struct {
		client     *ethclient.Client
		Users      map[int64]*User
		orders     map[int64]int64
		PrivateKey *ecdsa.PrivateKey
		orderbook  map[Market]*orderbook.Orderbook
	}

	MatchedOrder struct {
		Price float64
		Size  float64
		ID    int64
	}

	Order struct {
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderbookData struct {
		TotalBidVolumne float64
		TotalAskVolumne float64
		Asks            []*Order
		Bids            []*Order
	}

	User struct {
		Id         int64
		PrivateKey *ecdsa.PrivateKey
	}
)

// const decleration
const (
	MarketEth Market = "ETH"
	to        string = "0x1dF62f291b2E969fB0849d99D9Ce41e2F137006e"

	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
)

func NewUser(privateKey string) *User {

	prk, err := crypto.HexToECDSA(privateKey)

	if err != nil {
		panic(err)
	}

	return &User{
		PrivateKey: prk,
	}
}

func NewExchange(privateKey string, client *ethclient.Client) *Exchange {

	orderbooks := make(map[Market]*orderbook.Orderbook)

	orderbooks[MarketEth] = orderbook.NewOrderbook()

	prKey, err := crypto.HexToECDSA(privateKey)

	if err != nil {
		log.Fatal(err)
	}

	return &Exchange{
		client:     client,
		Users:      make(map[int64]*User),
		orders:     make(map[int64]int64),
		orderbook:  orderbooks,
		PrivateKey: prKey,
	}
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

func (ex *Exchange) handlePlaceMarketOrder(market Market, o *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {

	ob := ex.orderbook[market]
	matches := ob.PlaceMarketOrder(o)
	matchedOrders := make([]*MatchedOrder, len(matches))

	isBid := false

	if o.Bid {
		isBid = true
	}

	for i := 0; i < len(matchedOrders); i++ {
		var matchedID int64
		if isBid {
			matchedID = matches[i].Bid.ID // Assuming Bid has an ID field
		} else {
			matchedID = matches[i].Ask.ID // Assuming Ask has an ID field
		}

		matchedOrders[i] = &MatchedOrder{
			Size:  matches[i].SizeFilled,
			Price: matches[i].Prize,
			ID:    matchedID,
		}
	}
	return matches, matchedOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, o *orderbook.Order) error {

	ob := ex.orderbook[market]
	ob.PlaceLimitOrder(price, o)

	user, ok := ex.Users[o.UserID]

	if !ok {
		return fmt.Errorf("user not found: %d", o.UserID)
	}

	prk := user.PrivateKey

	// Convert the string address to common.Address
	toAddress := common.HexToAddress(to)

	// Call transferEth with the correct types
	return transferEth(ex.client, prk, toAddress, o.Size())

}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {

	// for _ , match := range matches {

	// }

	return nil

}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placemarkerorder PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placemarkerorder); err != nil {
		return err
	}

	fmt.Println(placemarkerorder)

	order := orderbook.NewOrder(placemarkerorder.Bid, placemarkerorder.Size, int64(placemarkerorder.UserId))

	if placemarkerorder.Type == OrderType(LimitOrder) {
		fmt.Println(order)
		if err := ex.handlePlaceLimitOrder(Market(placemarkerorder.Market), placemarkerorder.Price, order); err != nil {
			return err
		}
		return c.JSON(200, map[string]any{"msg": "limit order placed"})
	}

	if placemarkerorder.Type == OrderType(MarketOrder) {
		matches, matchedOrders := ex.handlePlaceMarketOrder(Market(placemarkerorder.Market), order)

		ex.handleMatches(matches)

		return c.JSON(200, map[string]any{"matches": matchedOrders})
	}

	return nil

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
