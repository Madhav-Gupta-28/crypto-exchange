package main

import (
	"time"

	"github.com/Madhav-Gupta-28/crypto-exchange/client"
	"github.com/Madhav-Gupta-28/crypto-exchange/server"
)

func main() {

	go server.StartServer()

	time.Sleep(1 * time.Second)

	client := client.NewClient()

	params := &client.PlaceLimitOrderParams{
		UserId: 69,
		Size:   1,
		Price:  100,
		Bid:    false,
		Market: "ETH",
		Type:   "LIMIT",
	}

	client.PlaceLimitOrder(params)

}
