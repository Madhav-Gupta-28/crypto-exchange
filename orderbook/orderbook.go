package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type (
	Match struct {
		Ask        *Order
		Bid        *Order
		SizeFilled float64
		Prize      float64
	}

	Order struct {
		UserID    int64
		ID        int64
		size      float64
		Bid       bool
		Limit     *Limit
		Timestamp int64
	}

	Orders []*Order

	Limit struct {
		Price       float64
		Orders      Orders
		TotalVolume float64
	}

	Limits []*Limit

	ByBestAsk struct{ Limits }

	Orderbook struct {
		ask  []*Limit
		bids []*Limit

		AskLimits map[float64]*Limit
		BidLimits map[float64]*Limit

		orders map[int64]*Order
	}
)

func (o Orders) Len() int           { return len(o) }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

func (o *Order) String() string {
	return fmt.Sprintf("size : [%.2f]", o.size)
}

func NewOrder(bid bool, size float64, userID int64) *Order {
	return &Order{
		UserID:    userID,
		ID:        int64(rand.Intn(100000)),
		size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

// For asks
func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

// For bids
type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }

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

func (l *Limit) String() string {
	return fmt.Sprintf("prize : [%.2f]  TotalVolumne : [%.2f] ", l.Price, l.TotalVolume)
}

func (l *Limit) DeleteOrder(o *Order) string {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			l.TotalVolume -= o.size
			return fmt.Sprintf("Order is removed %.2F", o.Limit.Price)
		}
	}
	o.Limit = nil

	sort.Sort(l.Orders)

	return "Order not found"
}

func (l *Limit) fillOrder(a *Order, b *Order) Match {

	var (
		bid        *Order
		ask        *Order
		SizeFilled float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if a.size >= b.size {
		a.size -= b.size
		SizeFilled = b.size
		b.size = 0
	} else {
		b.size -= a.size
		SizeFilled = a.size
		a.size = 0
	}

	return Match{bid, ask, SizeFilled, l.Price}
}

func (o *Order) isFilled() bool {

	if o.size == 0.0 {
		return true
	} else {
		return false
	}
}

func (l *Limit) Fill(o *Order) []Match {

	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, order := range l.Orders {
		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if order.isFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}

		if order.isFilled() {
			break
		}

	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches

}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		ask:       []*Limit{},
		bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		orders:    make(map[int64]*Order),
	}
}

func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {

	matches := []Match{}

	if o.Bid {

		if ob.AskTotalVolumne() < o.size {
			panic(("not enough volume to fill order in the orderbook"))
		}

		for _, limit := range ob.Asks() {
			limitmatches := limit.Fill(o)
			matches = append(matches, limitmatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}

		}

	} else {

		if ob.BidTotalVolumne() < o.size {
			panic(("not enough volume to fill order in the orderbook"))
		}

		for _, limit := range ob.Bids() {
			limitmatches := limit.Fill(o)
			matches = append(matches, limitmatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(false, limit)
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
		// limit.AddOrder(o)

		if o.Bid {
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.ask = append(ob.ask, limit)
			ob.AskLimits[price] = limit
		}
	}

	ob.orders[o.ID] = o
	limit.AddOrder(o)

}

func (ob *Orderbook) CancelOrder(o *Order) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.orders, o.ID)
}

func (ob *Orderbook) clearLimit(bid bool, l *Limit) {
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
		for i := 0; i < len(ob.ask); i++ {
			if ob.ask[i] == l {
				ob.ask[i] = ob.ask[len(ob.ask)-1]
				ob.ask = ob.ask[:len(ob.ask)-1]
			}
		}

	}

}

func (ob *Orderbook) BidTotalVolumne() float64 {
	totalVolumne := 0.0

	for i := 0; i < len(ob.bids); i++ {

		totalVolumne += ob.bids[i].TotalVolume

	}
	return totalVolumne

}

func (ob *Orderbook) AskTotalVolumne() float64 {
	totalVolumne := 0.0

	for i := 0; i < len(ob.ask); i++ {

		totalVolumne += ob.ask[i].TotalVolume

	}
	return totalVolumne

}

func (ob *Orderbook) Asks() []*Limit {

	sort.Sort(ByBestAsk{ob.ask})
	return ob.ask
}

// Add this getter method to the Order struct
func (o *Order) Size() float64 {
	return o.size
}

func (ob *Orderbook) Bids() []*Limit {

	sort.Sort(ByBestBid{ob.bids})
	return ob.bids

}
