package main

import (
	"fmt"
	"time"

	"github.com/Madhav-Gupta-28/crypto-exchange/client"
	"github.com/Madhav-Gupta-28/crypto-exchange/server"
)

func main() {

	go server.StartServer()

	time.Sleep(1 * time.Second)

	bidp := &client.PlaceOrderRequest{
		UserId: 9,
		Size:   10,
		Price:  10000,
		Bid:    true,
		Type:   server.LimitOrder,
	}

	// askpLimit := &client.PlaceOrderRequest{
	// 	UserId: 8,
	// 	Size:   7,
	// 	Price:  10000,
	// 	Bid:    false,
	// 	Type:   server.LimitOrder,
	// }

	go func() {
		for {
			response, err := client.NewClient().ClientPlaceLimitOrder(bidp)

			if err != nil {
				fmt.Println(err)
				continue
			}

			// response2, err2 := client.NewClient().ClientPlaceLimitOrder(askpLimit)
			// if err2 != nil {
			// 	fmt.Println(err2)
			// 	continue
			// }

			fmt.Println(response)
			// fmt.Println(response2)

			time.Sleep(5 * time.Second)

			client.NewClient().ClientCancelOrder(&client.CancelOrderRequest{
				OrderID: response.OrderID,
				Market:  "ETH",
			})

			bestBid, err := client.NewClient().ClientGetBestAsk(&client.GetBookRequest{
				Market: "ETH",
			})

			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("Best ask", bestBid)

			time.Sleep(2 * time.Second)

		}

	}()

	askp := &client.PlaceOrderRequest{
		UserId: 8,
		Size:   7,
		Price:  10000,
		Bid:    false,
		Type:   server.MarketOrder,
	}

	go func() {
		for {
			client.NewClient().ClientPlaceLimitOrder(askp)

			time.Sleep(2 * time.Second)

		}

	}()

	select {}

}
