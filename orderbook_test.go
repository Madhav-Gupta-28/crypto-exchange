package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)

	ordera := NewOrder(true, 10_000, l)
	orderb := NewOrder(true, 10_000, l)

	l.AddOrder(ordera)
	l.AddOrder(orderb)

	assert.Equal(t, len(l.Orders), 2)

	l.DeleteOrder(ordera)

	assert.Equal(t, len(l.Orders), 1)

}

func TestOrderbook(t *testing.T) {

	ob := NewOrderbook()

	buyorder := NewOrder(true, 10_000, nil)
	askorder := NewOrder(false, 10_000, nil)

	ob.PlaceOrder(10_000, buyorder)
	ob.PlaceOrder(10_000, askorder)

	fmt.Println(ob.Asks[0])

	// ob.PlaceMarketOrder(true, 10_000)
}
