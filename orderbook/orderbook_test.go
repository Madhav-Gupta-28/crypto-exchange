package orderbook

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)

	ordera := NewOrder(true, 10_000, 1)
	orderb := NewOrder(true, 10_000, 1)

	l.AddOrder(ordera)
	l.AddOrder(orderb)

	assert.Equal(t, len(l.Orders), 2)

	l.DeleteOrder(ordera)

	assert.Equal(t, len(l.Orders), 1)

}

func TestOrderbook(t *testing.T) {

	ob := NewOrderbook()

	buyorder := NewOrder(true, 10_000, 1)
	askorder := NewOrder(false, 10_000, 1)

	ob.PlaceLimitOrder(10_000, buyorder)
	ob.PlaceLimitOrder(10_000, askorder)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	buyorder := NewOrder(true, 10_000, 1)
	askorder := NewOrder(false, 10_000, 1)

	ob.PlaceLimitOrder(10_000, buyorder)
	ob.PlaceLimitOrder(10_000, askorder)

	assert.Equal(t, len(ob.bids), 1)
	assert.Equal(t, len(ob.asks), 1)

}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellorder := NewOrder(false, 20, 1)
	ob.PlaceLimitOrder(10_000, sellorder)

	buyorder := NewOrder(true, 20, 1)
	matches := ob.PlaceMarketOrder(buyorder)

	assert.Equal(t, len(matches), 1)
	assert.Equal(t, matches[0].SizeFilled, 20.0)
	assert.Equal(t, matches[0].Price, 10_000.0)

	fmt.Println(matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderbook()

	buyorderA := NewOrder(true, 20, 1)
	buyorderB := NewOrder(true, 20, 1)
	buyorderC := NewOrder(true, 20, 1)
	buyorderD := NewOrder(true, 1, 1)

	ob.PlaceLimitOrder(10_000, buyorderA)
	ob.PlaceLimitOrder(10_000, buyorderD)
	ob.PlaceLimitOrder(9_000, buyorderB)
	ob.PlaceLimitOrder(5_000, buyorderC)

	assert.Equal(t, ob.BidTotalVolumne(), 61.0)

	// sellorderA := NewOrder(false, 20)
	// matches := ob.PlaceMarketOrder(sellorderA)

	// assert.Equal(t, len(matches), 3)

}
