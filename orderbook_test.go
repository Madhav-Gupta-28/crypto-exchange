package main

import (
	"fmt"
	"testing"
)

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)

	buyorder := NewOrder(true, 100)
	buyorder1 := NewOrder(true, 200)

	l.AddOrder(buyorder)

	l.AddOrder(buyorder1)

	fmt.Println(l)

	// l.DeleteOrder(buyorder)

	// fmt.Println(l)

}

func TestOrderBook(t *testing.T) {

	ob1 := NewOrderbook()

	buyOrder := NewOrder(true, 10)

	ob1.PlaceOrder(18_000, buyOrder)

	fmt.Println(ob1.Bids)
}
