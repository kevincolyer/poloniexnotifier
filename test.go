package main

import (
	"fmt"
	"log"

	"gitlab.com/wmlph/poloniex-api"
)

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
mytrades, err := p.PrivateTradeHistoryAllWeek()
// 	mytrades, err := p.PrivateTradeHistory("BTC_ETH")
// 	mytrades, err := p.OpenOrders("BTC_ETH")
// mytrades, err := p.OpenOrdersAll()
        if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", mytrades)
        ticker,err:=p.Ticker()
        if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", ticker["USDT_BTC"])
        USDT_BTC:=ticker["USDT_BTC"].Last
        fmt.Println("Last price of Bitcoin was: ", USDT_BTC)
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
