package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"strings"
	"time"
        "github.com/pkg/errors"

	"gitlab.com/wmlph/poloniex-api"
	"gopkg.in/gomail.v2"
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

type MyOpenOrder struct {
	poloniex.OpenOrder
	CurrentRate float64
	Proximity   float64
	Pair        CurrencyPair
}
type MyOpenOrders []MyOpenOrder

type MyTradeEntry struct {
	poloniex.PrivateTradeHistoryEntry
	//	CurrentRate float64
	Pair      CurrencyPair
	TradeDate time.Time
}
type MyTradeHistory []MyTradeEntry

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

func (p CurrencyPair) Poloniex() string {
	return fmt.Sprintf("%s_%s", p.Base, p.Trade)
}

type Currency float64

func (c Currency) Length() int {
	return len(c.String())
}

func (c Currency) String() (s string) {
	i := strings.Split(fmt.Sprintf("%.9f", float64(c)), ".")
	s = everyThird(reverseStr(i[0]), ",")
	s = reverseStr(s) + "." + everyThird(i[1], " ")
	return
}

func Comma(n float64) string {
	i := strings.Split(fmt.Sprintf("%.2f", n), ".")
	return reverseStr(everyThird(reverseStr(i[0]), ",")) + "." + i[1]
}

func everyThird(str, insert string) (s string) {
	if str == "" {
		return
	}
	for len(str) > 0 {
		l := len(str)
		if l > 3 {
			if str[3] == '-' || str[3] == '+' {
				l = 4
			} else {
				l = 3
			}
		}
		s = s + str[:l]
		str = str[l:]
		if len(str) > 0 {
			s += insert
		}
	}
	return
}

func reverseStr(str string) (out string) {
	for _, s := range str {
		out = string(s) + out
	}
	return
}

type column struct {
	title  string
	width  int
	widthr int // used for DOT
	widthl int // used for DOT
	align  int
	dot    string "."
}

const (
	LEFT = iota + 1
	RIGHT
	CENTRE
	DOT
)

const poloniexTime = "2006-01-02 15:04:05"

type PrettyTable struct {
	columns []column
	rows    [][]string
	colsep  string
	rowsep  string
	padding string
	html    bool
	footer  int // 0 if none or row number the footer begins on
}

func NewPrettyTable() (t *PrettyTable) {
	t = new(PrettyTable)
	t.rowsep = "-"
	t.colsep = "|"
	t.html = false
	t.padding = " "
	return
}

func (t *PrettyTable) addColumn(c *column) *PrettyTable {
	c.width = len(c.title) // in case title is wider than data
	t.columns = append(t.columns, *c)
	return t
}

func (t *PrettyTable) addRow(cols []string) *PrettyTable {
	t.rows = append(t.rows, cols)
	// get max width as we add columns in. Dot centred is harder to calc and need left and right of decimal point widths taken into consideration.
	for i, c := range t.columns {
		if i >= len(cols) {
			break
		} // skip empty columns
		w := len(cols[i]) // current width of col in this row
		if w < 2+len(c.dot) {
			continue // must have enough chars to split
		}
		if c.align == DOT && strings.Contains(cols[i], c.dot) == true {
			j := strings.Split(cols[i], c.dot)
			if len(j) != 2 {
				continue
			} // must split only in two
			t.columns[i].widthl = max(t.columns[i].widthl, len(j[0]))
			t.columns[i].widthr = max(t.columns[i].widthr, len(j[1]))
			w = t.columns[i].widthl + t.columns[i].widthr + len(c.dot)
		}
		t.columns[i].width = max(t.columns[i].width, w)
	}
	return t
}

func (t *PrettyTable) addFooter(cols []string) *PrettyTable {
	tablelength := len(t.rows)
	if t.footer == 0 {
		t.footer = tablelength - 1
	}
	t.addRow(cols)
	return t
}

func (t *PrettyTable) String() (s string) {
	// print header
	var txt string
	pad := t.padding
	nl := "\n"
	// sep header sep
	emph := true
	for _, col := range t.columns {
		if len(txt) > 0 {
			txt += t.colsep
		}
		txt += pad + padcentre(col.title, col.width) + pad

	}
	if emph == true {
		txt = strings.ToUpper(txt)
	}
	emph = false
	bar := strings.Repeat(t.rowsep, len(txt)) + nl
	s = bar + txt + nl + bar
	// rows final sep
	for j, row := range t.rows {
		txt = ""
		for i, col := range row {
			if len(txt) > 0 {
				txt += t.colsep
			}
			txt += pad + t.columns[i].aligntext(col) + pad
		}
		if emph == true {
			txt = strings.ToUpper(txt)
		} // if in footer or header
		s += txt + nl
		// if just before the footer...
		if j == t.footer && j > 0 {
			s += bar
			emph = true
		} // in footer now...
	}
	s += bar
	if t.html {
		s = "<pre>" + s + "</pre>"
	}
	return
}

func (c column) aligntext(text string) (s string) {
	switch c.align {
	case LEFT:
		s = padl(text, c.width)
	case RIGHT:
		s = padr(text, c.width)
	case CENTRE:
		s = padcentre(text, c.width)
	case DOT:

		i := strings.Split(text, c.dot)
		if len(i) != 2 {
			s = padl(text, c.width)
			return s
		}
		s = padl(i[0], c.widthl) + c.dot + padr(i[1], c.widthr)
	default:
		s = "unforseen error"
	}
	return s
}

func padl(text string, width int) string {
	return spaces(width-len(text)) + text
}

func padr(text string, width int) string {
	return text + spaces(width-len(text))
}

func padcentre(text string, width int) string {
	add := spaces((len(text)%2 + width%2) % 2) // 1 if odd, 0 if even
	lr := spaces((width - len(text)) / 2)

	return lr + text + lr + add
}

func spaces(width int) string {
	if width < 0 { //fmt.Println("width is less than 0")
		width = 0
	}
	return strings.Repeat(" ", width)
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func heading(text string) string {
	return strings.ToUpper(text) + "\n" + strings.Repeat("-", len(text))
}

type Report struct {
	Emailto       string
	Emailfrom     string
	Smtpserver    string
	Reportname    string
	Port          float64
	Smtp_login    string
	Smtp_password string
	Smtp_ssl      bool
	Subject       string
	Body          string
	Topmovers     float64
}

// unmarshal only fills exported fields!!!

func NewReport(conf string) *Report {
	r := &Report{}
// 	fmt.Println(r)
	b, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "reading "+conf+" failed."))
	}
// 	fmt.Printf("%s\n",b)
	err = json.Unmarshal(b, r)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "unmarshalling json failed"))
	}
// 	fmt.Println(r)
	if r.Reportname=="" {
            log.Fatalln("Error - no report name specified in config file "+conf)
        }
        
        // defaults
	r.Subject = "Poloniex Activity Report for " + r.Reportname + " at " + time.Now().Format(poloniexTime)
        if r.Topmovers==0 { r.Topmovers = 15 }
	return r

}

func main() {
	p := poloniex.New("config.json")
	report := NewReport("reportconfig.json")

	report.Body = fmt.Sprintln(heading(report.Subject) + "\n\n")
	/*
	   Prices
	*/
	report.Body += fmt.Sprintln("\n" + heading("Prices"))

	ticker, err := p.Ticker()
	if err != nil {
		log.Fatalln(err)
	}

	//fmt.Printf("%+v\n", ticker["USDT_BTC"]) == {Last:4192.58071046 Ask:4192.58071046 Bid:4186.2093 Change:0.03316429 BaseVolume:3.534202327390681e+07 QuoteVolume:8405.26739006 IsFrozen:0}

	USDT_BTC := ticker["USDT_BTC"].Last
	report.Body += fmt.Sprintln("Last price of Bitcoin : $", Comma(USDT_BTC), fmt.Sprintf("(%+.0f%%)", ticker["USDT_BTC"].Change*100))
	report.Body += fmt.Sprintln("Last price of Ethereum: $", Comma(ticker["USDT_ETH"].Last), fmt.Sprintf("(%+.0f%%)", ticker["USDT_ETH"].Change*100))
        report.Body += fmt.Sprintln()

        /*
        Top Movers 
        */
        tm:=int(report.Topmovers)
	report.Body += fmt.Sprintln("\n" + heading(fmt.Sprintf("Top %v%% Movers",tm)))
        t:= NewPrettyTable()
//         report.Body
        type mover struct {
            name CurrencyPair
            rate float64
            change int
        }
        
        var movers []mover 
        
        for c,i:=range ticker {
            if int(math.Abs(i.Change)*100)<tm { continue }
            pair:=NewCurrencyPair(c)
            if pair.Base!="BTC" { continue }
            movers=append(movers,mover{ 
                name: pair, 
                rate: i.Last, 
                change: int(i.Change*100)} )
        }
        sort.Slice(movers,func(i,j int) bool {return movers[i].change<movers[j].change})
        
        t.addColumn(&column{title: "Currency", align: RIGHT})
        t.addColumn(&column{title: "Rate", align: DOT,dot: "."})
        t.addColumn(&column{title: "Change", align: DOT,dot: "."})

        for _,o:=range movers {
            i := fmt.Sprintf("%s|%v|%v%%", o.name, Currency(o.rate), o.change)
		t.addRow(strings.Split(i, "|"))
	}
        report.Body+= fmt.Sprintln(t)
        
	/*
	   recent trades
	*/

	report.Body += fmt.Sprintln("\n" + heading("Recent Trades"))

	// added a patch to poloniex api to provide the function below
	mytrades, err := p.PrivateTradeHistoryAllWeek()
	if err != nil {
		log.Fatalln(err)
	}
	//fmt.Printf("%+v\n", mytrades)
	/*
		  	t, _ := time.Parse(poloniex, "2017-09-06 16:32:11")
			fmt.Println(t)
			fmt.Println(t.Format(time.RFC850))

	*/

	var mytradehistory MyTradeHistory
	for curr, trades := range mytrades {
		for _, t := range trades {
			td, _ := time.Parse(poloniexTime, t.Date)
			mytradehistory = append(mytradehistory,
				MyTradeEntry{
					PrivateTradeHistoryEntry: t,
					Pair:      NewCurrencyPair(curr),
					TradeDate: td,
				},
			)
		}
	}
	sort.Slice(mytradehistory, func(i, j int) bool { return mytradehistory[i].Date > mytradehistory[j].Date })

	t= NewPrettyTable()
	//	t.html=true
	t.addColumn(&column{title: "24", align: RIGHT})
	t.addColumn(&column{title: "Order", align: LEFT})
	t.addColumn(&column{title: "Date", align: LEFT})
	t.addColumn(&column{title: "Type", align: LEFT})
	t.addColumn(&column{title: "Rate", align: DOT, dot: "."})
	t.addColumn(&column{title: "Amount", align: DOT, dot: "."})
	t.addColumn(&column{title: "Total", align: DOT, dot: "."})
	t.addColumn(&column{title: "Value", align: DOT, dot: "."})
	t.addColumn(&column{title: "24", align: LEFT})

	gain := 0.0
	p24h := ""
	past24hours := time.Now().Add(time.Duration(-24) * time.Hour)
	//pastweek := time.Now().Add(time.Duration(-24*7) * time.Hour)

	for _, o := range mytradehistory {
		//if o.TradeDate.Before(pastweek) { continue }
		k := o.Total
		if o.Pair.Base == "BTC" {
			k = USDT_BTC * k
		}
		if o.Type == "buy" {
			k = -k
		}
		gain += k
		if o.TradeDate.Before(past24hours) {
			p24h = " "
		} else {
			p24h = "*"
		}
		i := fmt.Sprintf("%s|%9s|%s|%-4s|%v|%v|%v|%v USDT|%s", p24h, o.Pair, o.Date, o.Type, Currency(o.Rate), Currency(o.Amount), Currency(o.Total), Comma(k), p24h)
		t.addRow(strings.Split(i, "|"))
	}
	t.addFooter([]string{"", "Net gain", "", "", "", "", "", Comma(gain) + " USDT", ""})
	report.Body += fmt.Sprintln(t)

	/*
		OpenOrders
	*/

	report.Body += fmt.Sprintln("\n" + heading("My Open Orders"))
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

	t = NewPrettyTable()
	t.addColumn(&column{title: "Order", align: LEFT})
	t.addColumn(&column{title: "Prox", align: LEFT})
	t.addColumn(&column{title: "Type", align: LEFT, dot: "."})
	t.addColumn(&column{title: "Rate", align: DOT, dot: "."})
	t.addColumn(&column{title: "Amount", align: DOT, dot: "."})
	t.addColumn(&column{title: "Value", align: DOT, dot: "."})
	t.addColumn(&column{title: "Gain", align: DOT, dot: "."})
	t.addColumn(&column{title: "24hrs", align: LEFT})
	gain = 0.0
	asnow := 0.0
	j := 0.0
	for _, o := range myorders {
		k := o.Total
		if o.Pair.Base == "BTC" {
			k = USDT_BTC * k
		}
		gain += k
		j = o.Amount
		if o.Pair.Base == "BTC" {
			j = j * ticker[o.Pair.Poloniex()].Last * USDT_BTC
		}
		asnow += j
		//
		i := fmt.Sprintf("%9s|%3.0f%%|%-4s|%v|%v|%v %s|$%v|%+.0f%%", o.Pair, o.Proximity, o.Type, Currency(o.Rate), Currency(o.Amount), Currency(o.Total), o.Pair.Base, Comma(k), ticker[o.Pair.Poloniex()].Change*100)

		t.addRow(strings.Split(i, "|"))
	}
	// and print the resulting table!
	t.addFooter([]string{"If realised", "", "", "", "", "", "$" + Comma(gain), ""})
	t.addFooter([]string{"As now", "", "", "", "", "", "$" + Comma(asnow), ""})
	t.addFooter([]string{"Profit", "", "", "", "", "", "$" + Comma(gain-asnow), ""})
	report.Body += fmt.Sprintln(t)

	/*
	   send message
	*/
	fmt.Println(report.Body)
	report.Send()
}

func (r *Report) Send() (e error) {
	m := gomail.NewMessage()
	m.SetHeader("From", r.Emailfrom)
	m.SetHeader("To", r.Emailto)
	m.SetHeader("Subject", r.Subject)
	m.SetBody("text/html", "<pre>"+r.Body+"</pre>")
// 	fmt.Println(m)
	d := gomail.NewDialer(r.Smtpserver, int(r.Port), r.Smtp_login, r.Smtp_password) // with or without auth
	//d := gomail.Dialer{Host: "127.0.0.1", Port: 25, SSL: false, Auth: nil} // no auth
	//if r.smtp_ssl {d.SSL=true}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
	return
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
