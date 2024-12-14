package main

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {

	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v ", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)

	buyorder := NewOrder(true, 100)
	buyorder1 := NewOrder(true, 200)

	l.AddOrder(buyorder)

	l.AddOrder(buyorder1)

	assert(t, len(l.Orders), 2)

	fmt.Println(l)

	// l.DeleteOrder(buyorder)

	// fmt.Println(l)

}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellorder := NewOrder(false, 10)

	ob.PlaceLimitOrder(10, sellorder)

	assert(t, len(ob.ask), 1)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellorder := NewOrder(false, 20)
	ob.PlaceLimitOrder(10000, sellorder)
	buyorder := NewOrder(true, 10)

	// ob.PlaceMarketOrder(sellorder)
	// ob.PlaceMarketOrder(buyorder)

	matches := ob.PlaceMarketOrder(buyorder)

	assert(t, len(matches), 1)

	assert(t, ob.AskTotalVolumne(), 10.0)

	// assert(t, matches[0].Ask, sellorder)
	// assert(t, matches[0].Bid, buyorder)

	fmt.Printf("%+v\n", matches)

	assert(t, len(ob.ask), 1)
}

func TestPlaceMarketOrderMultifill(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 10)
	buyOrderC := NewOrder(true, 15)
	buyOrderD := NewOrder(true, 20)

	ob.PlaceLimitOrder(1000, buyOrderA)
	ob.PlaceLimitOrder(9000, buyOrderB)
	ob.PlaceLimitOrder(11000, buyOrderC)
	ob.PlaceLimitOrder(1000, buyOrderD)

	assert(t, ob.BidTotalVolumne(), 50.0)
	// assert(t, ob.AskTotalVolumne(), 0.0)

	// sellOrder := NewOrder(false, 20)
	// matches := ob.PlaceMarketOrder(sellOrder)

	// fmt.Printf("%+v\n", matches)

	// assert(t, len(matches), 3)
	// assert(t, len(ob.bids), 3)

	// assert(t, ob.AskTotalVolumne(), 10.0)

}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 10)

	ob.PlaceLimitOrder(1000, buyOrderA)
	ob.PlaceLimitOrder(9000, buyOrderB)

	assert(t, ob.BidTotalVolumne(), 15.0)

	ob.cancelOrder(buyOrderB)

	assert(t, ob.BidTotalVolumne(), 5.0)

}
