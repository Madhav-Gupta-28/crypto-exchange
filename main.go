package main

import (
	"fmt"
	"math"
	"time"

	"github.com/Madhav-Gupta-28/crypto-exchange/client"
	"github.com/Madhav-Gupta-28/crypto-exchange/server"
)

const (
	maxOrders = 3
)

var tick = 2 * time.Second

var myAsks = make(map[float64]int64)
var myBids = make(map[float64]int64)

func marketOrderPlacer(Client *client.Client) error {
	ticker := time.NewTicker(tick)
	for {
		<-ticker.C

		marketSell := &client.PlaceLimitOrderParams{
			UserId: 1,
			Size:   2,
			Bid:    false,
		}
		err := Client.PlaceMarketOrder(marketSell)
		if err != nil {
			return err
		}

		marketbuy := &client.PlaceLimitOrderParams{
			UserId: 2,
			Size:   2,
			Bid:    true,
		}
		err = Client.PlaceMarketOrder(marketbuy)
		if err != nil {
			return err
		}
	}
}

func makeMarketSimple(Client *client.Client) error {
	ticker := time.NewTicker(tick)
	stradle := 100.0

	bestAsk := 0.0
	bestBid := 0.0

	for {
		<-ticker.C

		bestAsk, _ = Client.GetBestAskPrice(server.Market("ETH"))
		fmt.Println(bestAsk)
		fmt.Println(bestAsk)

		bestBid, _ = Client.GetBestBidPrice(server.Market("ETH"))
		fmt.Println(bestBid)

		spread := math.Abs(bestAsk - bestBid)
		fmt.Println(spread)

		if len(myBids) < maxOrders {
			// Place a bid limit order
			bidLimit := &client.PlaceLimitOrderParams{
				UserId: 2,
				Size:   2,
				Price:  bestBid + stradle,
				Bid:    true,
			}
			orderId, err := Client.PlaceLimitOrder(bidLimit)
			if err != nil {
				return err
			}
			myBids[bestBid+stradle] = orderId.OrderId
		}

		// Place an ask limit order
		if len(myAsks) < maxOrders {
			askLimit := &client.PlaceLimitOrderParams{
				UserId: 1,
				Size:   1,
				Price:  bestAsk - stradle,
				Bid:    false,
			}
			orderId, err := Client.PlaceLimitOrder(askLimit)
			if err != nil {
				return err
			}
			myAsks[bestAsk-stradle] = orderId.OrderId

			spread := math.Abs(bestAsk - bestBid)
			fmt.Println(spread)

		}

	}
}

func seedMarket(c *client.Client) error {
	ask := &client.PlaceLimitOrderParams{
		UserId: 1,
		Size:   7,
		Price:  100,
		Bid:    false,
	}

	bid := &client.PlaceLimitOrderParams{
		UserId: 2,
		Size:   7,
		Price:  10,
		Bid:    true,
	}

	_, err := c.PlaceLimitOrder(ask)
	if err != nil {
		return err
	}
	_, err = c.PlaceLimitOrder(bid)
	if err != nil {
		return err
	}
	return nil

}

func main() {

	go server.StartServer()

	time.Sleep(1 * time.Second)

	c := client.NewClient()

	err := seedMarket(c)
	if err != nil {
		fmt.Println(err)
	}

	go makeMarketSimple(c)

	time.Sleep(1 * time.Second)

	marketOrderPlacer(c)

	select {}

	// for {
	// 	params := &client.PlaceLimitOrderParams{
	// 		UserId: 69,
	// 		Size:   5,
	// 		Price:  100,
	// 		Bid:    false,
	// 	}

	// 	_, err := c.PlaceLimitOrder(params)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}

	// 	time.Sleep(1 * time.Second)

	// 	params = &client.PlaceLimitOrderParams{
	// 		UserId: 79,
	// 		Size:   1,
	// 		Price:  100,
	// 		Bid:    true,
	// 	}

	// 	err = c.PlaceMarketOrder(params)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}

	// 	bestAskPrice, err := c.GetBestAskPrice(server.Market("ETH"))
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}

	// 	fmt.Println(bestAskPrice)

	// }

	//select {}

}
