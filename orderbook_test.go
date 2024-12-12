package main

import (
	"fmt"
	"testing"
)

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)

	buyorder := NewOrder(true, 100)

	l.AddOrder(buyorder)

	fmt.Println(l)

	l.DeleteOrder(buyorder)

	fmt.Println(l)

}

func TestOrderBook(t *testing.T) {
}
