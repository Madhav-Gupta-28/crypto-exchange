package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	"github.com/Madhav-Gupta-28/crypto-exchange/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
		client *ethclient.Client
		Users  map[int64]*User
		mu     sync.RWMutex

		// orders maps a user to it's orders
		Orders     map[int64][]*orderbook.Order
		orderbooks map[Market]*orderbook.Orderbook
		PrivateKey *ecdsa.PrivateKey
	}

	OrderResponse struct {
		UserId    int64
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
		UserId int64
		Price  float64
		Size   float64
		Id     int64
	}

	User struct {
		Id         int64
		PrivateKey *ecdsa.PrivateKey
	}

	PlaceOrderResponse struct {
		OrderId int64
	}

	BestBidResponse struct {
		Price float64
	}

	OrdersByUserIdResponse struct {
		Asks []*orderbook.Order
		Bids []*orderbook.Order
	}
)

func StartServer() {

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

	pv2, err2 := crypto.HexToECDSA("6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c")
	if err2 != nil {
		log.Fatal(err2)
	}

	user1 := &User{
		Id:         1,
		PrivateKey: pv1,
	}
	ex.Users[user1.Id] = user1

	user2 := &User{
		Id:         2,
		PrivateKey: pv2,
	}
	ex.Users[user2.Id] = user2

	// Get addresses for both users
	user1Address := crypto.PubkeyToAddress(*user1.PrivateKey.Public().(*ecdsa.PublicKey))
	user2Address := crypto.PubkeyToAddress(*user2.PrivateKey.Public().(*ecdsa.PublicKey))

	// Get balances
	user1Balance, err := client.BalanceAt(context.Background(), user1Address, nil)
	if err != nil {
		log.Fatal(err)
	}
	user2Balance, err := client.BalanceAt(context.Background(), user2Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User 1 (%s) balance: %s ETH\n", user1Address.Hex(),
		new(big.Float).Quo(new(big.Float).SetInt(user1Balance), big.NewFloat(1e18)))
	fmt.Printf("User 2 (%s) balance: %s ETH\n", user2Address.Hex(),
		new(big.Float).Quo(new(big.Float).SetInt(user2Balance), big.NewFloat(1e18)))

	e.POST("/order", ex.handlePlaceOrder)
	e.GET("/order/:userId", ex.handleGetOrdersByUserid)
	e.GET("/trades/:market", ex.handleGetTrades)
	e.GET("/book/:market", ex.handleGetOrderbook)
	e.GET("/book", ex.handleGetBook)
	e.DELETE("/order/:orderID", ex.handleCancelOrder)
	e.GET("/book/:market/bid", ex.handleGetBestBid)
	e.GET("/book/:market/ask", ex.handleGetBestAsk)

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
		Orders:     make(map[int64][]*orderbook.Order),
		orderbooks: make(map[Market]*orderbook.Orderbook),
		PrivateKey: pv,
	}
	ex.orderbooks[MarketETH] = orderbook.NewOrderbook()
	return ex
}

func NewUser(privateKey string, id int64) *User {
	pv, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatal(err)
	}
	return &User{
		Id:         id,
		PrivateKey: pv,
	}
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchedOrders := make([]*MatchedOrder, len(matches))

	isBid := false
	if order.Bid {
		isBid = true
	}

	totalFilled := 0.0
	sumPrice := 0.0
	for i := 0; i < len(matches); i++ {
		var limitUserId int64
		limitUserId = matches[i].Bid.UserId

		id := matches[i].Bid.Id
		if !isBid {
			id = matches[i].Ask.Id
			limitUserId = matches[i].Ask.UserId
		}
		matchedOrders[i] = &MatchedOrder{
			UserId: limitUserId,
			Price:  matches[i].Price,
			Size:   matches[i].SizeFilled,
			Id:     id,
		}
		totalFilled += matches[i].SizeFilled
		sumPrice += matches[i].Price * matches[i].SizeFilled
	}

	avgPrice := sumPrice / float64(len(matches))

	fmt.Printf("Average Price: %.2f\n", avgPrice)

	newOrdermap := make(map[int64][]*orderbook.Order)

	ex.mu.Lock()

	for userid, Orderbookorders := range ex.Orders {
		for i := 0; i < len(Orderbookorders); i++ {
			if !Orderbookorders[i].IsFilled() {
				newOrdermap[userid] = append(newOrdermap[userid], Orderbookorders[i])
			}
		}

	}

	ex.Orders = newOrdermap
	ex.mu.Unlock()

	return matches, matchedOrders
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	for _, match := range matches {
		// Determine who sends the ETH based on the match
		var fromUser, toUser *User
		var ok bool

		// If it's a sell order (Ask), seller sends ETH to buyer
		// If it's a buy order (Bid), buyer sends ETH to seller
		if match.Bid.Bid {
			// It's a buy order, buyer (Bid) sends ETH to seller (Ask)
			fromUser, ok = ex.Users[match.Bid.UserId]
			if !ok {
				return fmt.Errorf("buyer user not found")
			}
			toUser, ok = ex.Users[match.Ask.UserId]
			if !ok {
				return fmt.Errorf("seller user not found")
			}
		} else {
			// It's a sell order, seller (Ask) sends ETH to buyer (Bid)
			fromUser, ok = ex.Users[match.Ask.UserId]
			if !ok {
				return fmt.Errorf("seller user not found")
			}
			toUser, ok = ex.Users[match.Bid.UserId]
			if !ok {
				return fmt.Errorf("buyer user not found")
			}
		}

		toPubKey := toUser.PrivateKey.Public()
		toPubKeyECDSA, ok := toPubKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		}

		err := TransferETH(ex.client, fromUser.PrivateKey,
			crypto.PubkeyToAddress(*toPubKeyECDSA).Hex(),
			match.SizeFilled)
		if err != nil {
			return fmt.Errorf("transfer failed: %w", err)
		}
	}
	return nil
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)
	ex.mu.Lock()
	ex.Orders[order.UserId] = append(ex.Orders[order.UserId], order)
	ex.mu.Unlock()
	fmt.Printf("new Limit Order Placed [ %.2f] | size [%.2f]", order.Limit.Price, order.Size)
	return nil
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placeorderdata PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeorderdata); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	market := Market(placeorderdata.Market)
	order := orderbook.NewOrder(placeorderdata.Bid, placeorderdata.Size, placeorderdata.UserId)

	// LIMIT ORDER
	if placeorderdata.Type == LIMITORDER {
		err := ex.handlePlaceLimitOrder(market, placeorderdata.Price, order)

		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
	}

	// MARKET ORDER
	if placeorderdata.Type == MARKETORDER {
		matches, _ := ex.handlePlaceMarketOrder(market, order)

		err := ex.handleMatches(matches)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

	}

	resp := &PlaceOrderResponse{
		OrderId: order.Id,
	}
	return c.JSON(http.StatusOK, resp)
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
				UserId:    orders.UserId,
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
				UserId:    orders.UserId,
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

func (ex *Exchange) handleGetBestBid(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "Orderbook of This Market not found"})
	}

	if len(ob.Bids()) == 0 {
		return c.JSON(http.StatusOK, map[string]any{"message": "No bids found"})
	}
	bestBidPrice := ob.Bids()[0].Price
	resp := BestBidResponse{
		Price: bestBidPrice,
	}
	return c.JSON(http.StatusOK, resp)
}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "Orderbook of This Market not found"})
	}

	if len(ob.Asks()) == 0 {
		return c.JSON(http.StatusOK, map[string]any{"message": "No bids found"})
	}
	bestBidPrice := ob.Asks()[0].Price
	resp := BestBidResponse{
		Price: bestBidPrice,
	}
	return c.JSON(http.StatusOK, resp)
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	return c.JSON(http.StatusOK, ex.orderbooks)
}

// Convert ETH to Wei
func EthToWei(eth float64) *big.Int {
	wei := new(big.Float).Mul(big.NewFloat(eth), big.NewFloat(1e18))
	weiInt, _ := wei.Int(nil) // Convert to big.Int
	return weiInt
}

func TransferETH(client *ethclient.Client, from *ecdsa.PrivateKey, to string, amount float64) error {
	// Convert ETH to Wei
	value := EthToWei(amount)

	// Since we already have the private key, we can use it directly
	publicKey := from.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	toAddress := common.HexToAddress(to)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID := big.NewInt(1337)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), from)
	if err != nil {
		return err
	}

	return client.SendTransaction(context.Background(), signedTx)

}

func (ex *Exchange) handleGetOrdersByUserid(c echo.Context) error {
	userIdStr := c.Param("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	// ex.mu.RLock()

	orderbooksOrders := ex.Orders[int64(userId)]
	if orderbooksOrders == nil {
		return c.JSON(http.StatusOK, []*OrderResponse{}) // Return empty array instead of null
	}

	orders := make([]*OrderResponse, 0, len(orderbooksOrders))

	for _, order := range orderbooksOrders {
		orderResp := &OrderResponse{
			Id:        order.Id,
			UserId:    order.UserId,
			Size:      order.Size,
			Bid:       order.Bid,
			TimeStamp: order.TimeStamp,
		}

		// Only set price if Limit exists
		if order.Limit != nil {
			orderResp.Price = order.Limit.Price
		}

		orders = append(orders, orderResp)
	}

	// ex.mu.RUnlock()

	return c.JSON(http.StatusOK, orders)
}

func (ex *Exchange) handleGetTrades(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "Orderbook of This Market not found"})
	}
	return c.JSON(http.StatusOK, ob.Trades)
}
