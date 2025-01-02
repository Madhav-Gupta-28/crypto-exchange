package main

import (
	"fmt"
	"sort"
	"time"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	Price      float64
	SizeFilled float64
}

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	TimeStamp int64
}

type Orders []*Order

func (o Orders) Len() int           { return len(o) }
func (o Orders) Less(i, j int) bool { return o[i].TimeStamp < o[j].TimeStamp }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

type Limit struct {
	Price        float64
	Orders       Orders
	TotalVolumne float64
}

type Orderbook struct {
	Asks []*Limit
	Bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

type Limits []*Limit
type ByBestAsk struct {
	Limits
}

func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }

type ByBestBid struct {
	Limits
}

func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }

// NewOrderbook creates a new orderbook
func NewOrderbook() *Orderbook {
	return &Orderbook{
		Asks:      []*Limit{},
		Bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

// NewLimitOrder creates a new limit order
func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

// NewOrder creates a new order
func NewOrder(bid bool, size float64, limit *Limit) *Order {
	return &Order{
		Size:      size,
		Limit:     limit,
		Bid:       bid,
		TimeStamp: time.Now().UnixNano(),
	}
}

// AddOrder adds an order to the limit
func (l *Limit) AddOrder(o *Order) {
	o.Limit = l                    // set the limit of the order
	l.Orders = append(l.Orders, o) // adding the order to the limit slice
	l.TotalVolumne += o.Size       // adding the order size to the total volume
}

func (l *Limit) DeleteOrder(o *Order) {
	l.TotalVolumne -= o.Size // subtracting the order size from the total volume
	// Here We are looping throuugh the order slice and the pushing out that order from the slice of Order in Limit
	for i, order := range l.Orders {
		if order == o {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			break
		}
	}
	o.Limit = nil // set the limit of the order to nil

	/// Sorting the orders in the limit
	sort.Sort(l.Orders)
}

func (o *Order) String() string {
	return fmt.Sprintf("[size] : %2.f", o.Size)
}

func (l *Limit) String() string {
	return fmt.Sprintf("[price] : %2.f and [total volume] : %2.f", l.Price, l.TotalVolumne)
}

func (ob *Orderbook) PlaceOrder(price float64, o *Order) []Match {

	// 1. Try to match the order

	// Add the order ot the orderbook
	if o.Size > 0 {
		ob.add(price, o)
	}

	return []Match{}

}

func (ob *Orderbook) add(price float64, o *Order) {

	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {

		limit = NewLimit(price)
		limit.AddOrder(o)
		if o.Bid {
			ob.Bids = append(ob.Bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.Asks = append(ob.Asks, limit)
			ob.AskLimits[price] = limit
		}
	}

}
