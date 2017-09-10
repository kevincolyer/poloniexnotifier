package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"gitlab.com/wmlph/poloniex-api"
)

type price float64
type amount float64
type Order poloniex.WSOrder

func (p price) String() (s string) {
	s = fmt.Sprintf("%0.8d", p)
	return
}

func (a amount) String() (s string) {
	s = fmt.Sprintf("%0.8d", a)
	return
}

func (o Order) String() (s string) {
	s = fmt.Sprintf("Type: %v Rate: %v Amount: %v", o.Type, o.Rate, o.Amount)
	return
}

// func (t PrivateTradeHistoryEntry) String() string {
//     fmt.Sprintf("(%s) Trade: ",t.Date,t.Type, )
// }
/*

	PrivateTradeHistoryEntry struct {
		Date        string
		Rate        float64 `json:",string"`
		Amount      float64 `json:",string"`
		Total       float64 `json:",string"`
		OrderNumber int64   `json:",string"`
		Type        string
	}*/

type MyOpenOrders []MyOpenOrder
type MyOpenOrder struct {
	poloniex.OpenOrder
	CurrentRate float64
	Proximity   float64
	Pair        CurrencyPair
}

func (o MyOpenOrder) String() string {
	return fmt.Sprintf("Prox.: %%%.0f %-4s Rate: %.9f Amount: %.9f Total: %.9f %s", o.Proximity, o.Type, o.Rate, o.Amount, o.Total, o.Pair.Base)
}

type CurrencyPair struct {
	Base  string
	Trade string
}

func NewCurrencyPair(s string) CurrencyPair {
	i := strings.Split(s, "_")
	var p CurrencyPair
	p.Base = i[0]
	if len(i) > 1 {
		p.Trade = i[1]
	}
	return p
}

func (p CurrencyPair) String() string {
	return fmt.Sprintf("%-4s/%4s", p.Trade, p.Base)
}

type Currency float64

func (c Currency) Length() int {
	return len(c.String())
}

func (c Currency) String() (s string) {
	i := strings.Split(fmt.Sprintf("%.9f", float64(c)), ".")
	s = everyThird(reverseStr(i[0]), ",")
	s = reverseStr(s) + "." + everyThird(i[1], "_")

	return
}

func everyThird(str, insert string) (s string) {
	s = ""
	for len(str) > 0 {
		l := len(str)
		if l > 3 {
			l = 3
		}
		s = s + str[:l]
		str = str[l:]
		if len(str) > 0 {
			s += insert
		}
		//         fmt.Printf("%s|%s\n",s,str)
	}
	return
}

func reverseStr(str string) (out string) {
	for _, s := range str {
		out = string(s) + out
	}
	return
}

func main() {

	p := poloniex.New("config.json")
	//  	balances, err := p.Balances()
	//                 if err != nil {
	// 		log.Fatalln(err)
	// 	}
	// 	fmt.Printf("%+v\n", balances)
	// my Trades
	// 	fmt.Println("My Trades")
	// 	mytrades, err := p.PrivateTradeHistoryAllWeek()
	// 	// 	mytrades, err := p.PrivateTradeHistory("BTC_ETH")
	// 	// 	mytrades, err := p.OpenOrders("BTC_ETH")
	// 	//      mytrades, err := p.OpenOrdersAll()
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}
	// 	fmt.Printf("%+v\n", mytrades)

	// Prices
	ticker, err := p.Ticker()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("\nPrices")
	fmt.Printf("%+v\n", ticker["USDT_BTC"])
	USDT_BTC := ticker["USDT_BTC"].Last
	fmt.Println("Last price of Bitcoin was: ", USDT_BTC)

	// 	// Balances
	// 	fmt.Println("\nBalances")
	// 	balances, err := p.Balances()
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}
	// 	fmt.Printf("%+v\n", balances)

	// my open OpenOrdersAll
	fmt.Println("\nMy Open Orders")
	openorders, err := p.OpenOrdersAll()
	if err != nil {
		log.Fatalln(err)
	}
	var myorders MyOpenOrders

	for pair, orders := range openorders {
		if len(orders) == 0 {
			continue
		}
		for _, order := range orders {
			myorder := MyOpenOrder{OpenOrder: order}
			//if myorder.Type=="Sell"
			myorder.Pair = NewCurrencyPair(pair)
			myorder.CurrentRate = ticker[pair].Bid
			diff := myorder.Rate - myorder.CurrentRate
			myorder.Proximity = diff / myorder.CurrentRate * 100
			myorders = append(myorders, myorder)
		}
	}
	//Sort by absolute value of proximity percentage ascending
	sort.Slice(myorders, func(i, j int) bool { return math.Abs(myorders[i].Proximity) < math.Abs(myorders[j].Proximity) })
	fmt.Printf("%-9s | %4s | %4s | %15s | %15s | %20s\n", "Order", "Prox", "Type", "Rate", "Amount", "Total")
	for _, o := range myorders {
		// 		fmt.Printf("Pair: %9s %v\n", order.Pair, order)
		fmt.Printf("%9s | %%%3.0f | %-4s | %-5.9f | %-5.9f | %-5.9f %s\n", o.Pair, o.Proximity, o.Type, o.Rate, o.Amount, o.Total, o.Pair.Base)
	}
	fmt.Printf("test %v \n", Currency(10000.123123123123))
	//fmt.Printf("%+v\n", openorders)
}

/* intialise
 * get price for btc/USDT_BTC
 * get PrivateTradeHistoryAll (perhaps extend with seen field)
 * are there any new trades?
 *  yes - print them
 *  email summary if requested
 * cache any values needed?
 *
 * config has email address
 * config has time between calls
 * config has secret and id
 *
 * tradedata: map of USDT_BTC, lastchecked,seentrades(id's)
 *
 * backed by tradedata.json
 *
 * GetNewTrades accepts start(time.Now) and prev time returns a slice of trades
 * commify
 * btc2usd
 * PurgeOldTrades acts on on tradedata and removes anything > 24hrs
 * MergeNewTrades acts on PrivateTradeHistoryAll and adds only those not yet seen
 */
