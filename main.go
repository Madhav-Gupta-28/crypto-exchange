package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/Madhav-Gupta-28/crypto-exchange/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)

func main() {
	fmt.Println("Hello world")

	// Echo Instance
	e := echo.New()

	e.HTTPErrorHandler = httpErrorHandler

	ex := NewExchange()

	e.GET("/book", ex.handleGetBook)

	e.DELETE("/cancel-order", ex.cancelOrder)

	e.POST("/place-order", ex.handlePlaceOrder)

	client, err := ethclient.Dial("http://localhost:8545")

	if err != nil {
		log.Fatal(err)
	}

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress("0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1"), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(balance)

	// Transferring Eth
	privateKey, err := crypto.HexToECDSA("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1000000000000000000) // in wei (1 eth)

	gasLimit := uint64(21000) // in units

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0xd03ea8624C8C5987235048901fB614fDcA89b117")

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	// chainID, err := client.NetworkID(context.Background())
	// fmt.Println(chainID)

	if err != nil {
		log.Fatal(err)
	}

	chainID := big.NewInt(1337)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex()) // tx sent

	e.Start(":3000")

}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

type Market string

const (
	MarketEth Market = "ETH"
)

type Exchange struct {
	PrivateKey *ecdsa.PrivateKey
	orderbook  map[Market]*orderbook.Orderbook
}

func NewExchange(privateKey string) *Exchange {

	orderbooks := make(map[Market]*orderbook.Orderbook)

	orderbooks[MarketEth] = orderbook.NewOrderbook()

	prKey, err := crypto.HexToECDSA(privateKey)

	if err != nil {
		log.Fatal(err)
	}

	return &Exchange{
		orderbook:  orderbooks,
		PrivateKey: prKey,
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

type MatchedOrder struct {
	Price float64
	Size  float64
	ID    int64
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
		matchedOrders := make([]*MatchedOrder, len(matches))

		isBid := false

		if order.Bid {
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

		return c.JSON(200, map[string]any{"matches": matchedOrders})
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
