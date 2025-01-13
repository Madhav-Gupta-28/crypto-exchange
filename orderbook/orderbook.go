package orderbook

import (
	"crypto/rand"
	"fmt"
	"math/big"
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
	Id        int64
	UserId    int64
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
	asks []*Limit
	bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
	Orders    map[int64]*Order
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
		asks:      []*Limit{},
		bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		Orders:    make(map[int64]*Order),
	}
}

func (ob *Orderbook) ClearLimit(bid bool, l *Limit) {

	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == l {
				ob.bids[i] = ob.bids[len(ob.bids)-1]
				ob.bids = ob.bids[:len(ob.bids)-1]
			}
		}
	} else {
		delete(ob.AskLimits, l.Price)
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == l {
				ob.asks[i] = ob.asks[len(ob.asks)-1]
				ob.asks = ob.asks[:len(ob.asks)-1]
			}
		}
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
func NewOrder(bid bool, size float64, userid int64) *Order {
	id, _ := rand.Int(rand.Reader, big.NewInt(1000000000000000000))
	return &Order{
		Id:        id.Int64(),
		UserId:    userid,
		Size:      size,
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

func (o *Order) isFilled() bool {
	return o.Size == 0
}

func (l *Limit) FillOrder(a, b *Order) Match {
	var (
		bid        *Order
		ask        *Order
		sizefilled float64
	)
	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}
	if a.Size > b.Size {
		a.Size -= b.Size
		sizefilled = b.Size
		b.Size = 0
	} else {
		b.Size -= a.Size
		sizefilled = a.Size
		a.Size = 0
	}
	return Match{
		Bid:        bid,
		Ask:        ask,
		Price:      l.Price,
		SizeFilled: sizefilled,
	}
}

func (l *Limit) Fill(o *Order) []Match {
	matches := []Match{}
	var ordersToDelete []*Order

	for _, order := range l.Orders {
		match := l.FillOrder(order, o)
		matches = append(matches, match)
		l.TotalVolumne -= match.SizeFilled

		if order.isFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}

		if o.isFilled() {
			break
		}
	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches

}

// CancelOrder cancels an order
func (ob *Orderbook) CancelOrder(o *Order) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.Orders, o.Id)
}

func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	matches := []Match{}

	if o.Bid {
		if ob.AskTotalVolumne() < o.Size {
			panic("not enough volume to fill the order")
		}
		for _, limit := range ob.Asks() {
			limitmatches := limit.Fill(o)
			matches = append(matches, limitmatches...)

			if len(limit.Orders) == 0 {
				ob.ClearLimit(true, limit)
			}
		}
	} else {
		if ob.BidTotalVolumne() < o.Size {
			panic("not enough volume to fill the order")
		}
		for _, limit := range ob.Bids() {
			limitmatches := limit.Fill(o)
			matches = append(matches, limitmatches...)

			if len(limit.Orders) == 0 {
				ob.ClearLimit(false, limit)
			}
		}
	}

	return matches

}

func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		if o.Bid {
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}
	limit.AddOrder(o)
	ob.Orders[o.Id] = o
}

// Asks returns the asks in the orderbook
func (ob *Orderbook) Asks() []*Limit {

	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

// Bids returns the bids in the orderbook
func (ob *Orderbook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}

// BidTotalVolumne returns the total volume of the bids in the orderbook
func (ob *Orderbook) BidTotalVolumne() float64 {
	total := 0.0
	for _, limit := range ob.bids {
		total += limit.TotalVolumne
	}
	return total
}

// AskTotalVolumne returns the total volume of the asks in the orderbook
func (ob *Orderbook) AskTotalVolumne() float64 {
	total := 0.0
	for _, limit := range ob.asks {
		total += limit.TotalVolumne
	}
	return total
}
