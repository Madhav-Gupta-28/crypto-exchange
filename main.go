package main

import (
	"time"

	"github.com/Madhav-Gupta-28/crypto-exchange/client"
	"github.com/Madhav-Gupta-28/crypto-exchange/server"
)

func main() {

	go server.StartServer()

	time.Sleep(1 * time.Second)

	c := client.NewClient()

	params := &client.PlaceLimitOrderParams{
		UserId: 69,
		Size:   1,
		Price:  100,
		Bid:    false,
	}

	c.PlaceLimitOrder(params)

	time.Sleep(15 * time.Second)

	params = &client.PlaceLimitOrderParams{
		UserId: 79,
		Size:   1,
		Price:  100,
		Bid:    true,
	}

	c.PlaceMarketOrder(params)

	select {}

}
