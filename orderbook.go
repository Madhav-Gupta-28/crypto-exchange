package main

import (
	"fmt"
	"time"
)

type Order struct {
	size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

func (o *Order) String() string {
	return fmt.Sprintf("size : [%.2f]", o.size)
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

type Limit struct {
	Price       float64
	Orders      []*Order
	TotalVolume float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
		// TotalVolume:  ,

	}

}

func (l *Limit) AddOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.size

}

func (l *Limit) DeleteOrder(o *Order) string {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			return fmt.Sprintf("Order is removed %.2F", o.Limit.Price)
		}
	}
	o.Limit = nil
	l.TotalVolume -= o.size

	return "Order not found"
}

type Orderbook struct {
	Asks []*Limit
	Bids []*Limit
}
