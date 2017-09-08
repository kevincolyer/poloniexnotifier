package main

import (
	"fmt"
	"log"

	"gitlab.com/wmlph/poloniex-api"
)

type price float64
type amount float64
type Order poloniex.WSOrder

func (p price) String() (s string) {
    s=fmt.Sprintf("%0.8d",p)
    return 
}

func (a amount) String() (s string) {
    s=fmt.Sprintf("%0.8d",a)
    return
}

func (o Order) String() (s string) {
    s=fmt.Sprintf("Type: %v Rate: %v Amount: %v",o.Type,o.Rate,o.Amount)
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
	
func main() {
    
    
    
	p := poloniex.New("config.json")
//  	balances, err := p.Balances()
//                 if err != nil {
// 		log.Fatalln(err)
// 	}
// 	fmt.Printf("%+v\n", balances)
        // my Trades
        fmt.Println("My Trades")
        mytrades, err := p.PrivateTradeHistoryAllWeek()
// 	mytrades, err := p.PrivateTradeHistory("BTC_ETH")
// 	mytrades, err := p.OpenOrders("BTC_ETH")
//      mytrades, err := p.OpenOrdersAll()
        if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", mytrades)
        
	// Prices
        ticker,err:=p.Ticker()
        if err != nil {
		log.Fatalln(err)
	}
        fmt.Println("\nPrices")
	fmt.Printf("%+v\n", ticker["USDT_BTC"])
        USDT_BTC:=ticker["USDT_BTC"].Last
        fmt.Println("Last price of Bitcoin was: ", USDT_BTC)
        
        // Balances
        fmt.Println("\nBalances")
       	balances, err := p.Balances()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", balances)

        // my open OpenOrdersAll
        fmt.Println("\nMy Open Orders")
        openorders, err := p.OpenOrdersAll()
         if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", openorders)
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
