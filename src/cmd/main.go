package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	moex "github.com/SeminDM/market/analyzer/moex"
	share "github.com/SeminDM/market/analyzer/share"
)

var shareTickers = []string{"PHOR", "SIBN", "ROSN", "SBER", "PLZL", "BELU"}
var currencyTickers = []string{"GLDRUB_TOM", "CNYRUB_TOM", "USD000UTSTOM"}

var colorReset = "\033[0m"
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var rowSeparator = "----------------------------------------------------------------------------------------------"

const issSecuritiesUri string = "https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities.xml?iss.meta=off&iss.only=marketdata,securities&marketdata.columns=SECID,LAST,VALTODAY&securities.columns=SECID,PREVPRICE"
const issIndexUri string = "https://iss.moex.com/iss/engines/stock/markets/index/boards/SNDX/securities.xml?iss.meta=off&iss.only=marketdata&marketdata.columns=SECID,LASTVALUE,CURRENTVALUE,VALTODAY&securities=IMOEX,RGBI"
const issRtsiUri string = "https://iss.moex.com/iss/engines/stock/markets/index/boards/RTSI/securities.xml?iss.meta=off&iss.only=marketdata&marketdata.columns=SECID,LASTVALUE,CURRENTVALUE,VALTODAY"
const issCurrencyUri string = "https://iss.moex.com/iss/engines/currency/markets/selt/securities.xml?iss.meta=off&iss.only=marketdata,securities&securities=CETS:USD000UTSTOM,CETS:GLDRUB_TOM,CETS:CNYRUB_TOM"

func main() {
	for i := 0; i < 1000; i++ {
		var doc moex.IssDocument
		if err := loadData(issSecuritiesUri, &doc); err != nil {
			panic(err)
		}
		sharesByTicker := make(map[string]*share.Share)
		for _, ticker := range shareTickers {
			s := share.New(ticker)
			sharesByTicker[ticker] = &s
		}
		if err := populateSecurities(sharesByTicker, doc.Data[1]); err != nil {
			panic(err)
		}
		if err := populateMarketData(sharesByTicker, doc.Data[0]); err != nil {
			panic(err)
		}
		shares := make([]*share.Share, 0, len(sharesByTicker))
		for _, v := range sharesByTicker {
			shares = append(shares, v)
		}
		sort.Sort(share.ByChange(shares))

		if err := loadData(issIndexUri, &doc); err != nil {
			panic(err)
		}
		imoex, err := getIndexData(doc.Data[2], "IMOEX")
		if err != nil {
			panic(err)
		}
		rgbi, err := getIndexData(doc.Data[2], "RGBI")
		if err != nil {
			panic(err)
		}

		if err := loadData(issRtsiUri, &doc); err != nil {
			panic(err)
		}
		rtsi, err := getIndexData(doc.Data[3], "RTSI")
		if err != nil {
			panic(err)
		}

		if err := loadData(issCurrencyUri, &doc); err != nil {
			panic(err)
		}

		currencyByTicker := make(map[string]*share.Share)
		for _, ticker := range currencyTickers {
			s := share.New(ticker)
			currencyByTicker[ticker] = &s
		}
		if err := populateSecurities(currencyByTicker, doc.Data[5]); err != nil {
			panic(err)
		}
		if err := populateMarketData(currencyByTicker, doc.Data[4]); err != nil {
			panic(err)
		}
		currencies := make([]*share.Share, 0, len(currencyByTicker))
		for _, v := range currencyByTicker {
			currencies = append(currencies, v)
		}
		sort.Sort(share.ByChange(currencies))

		printFrame(shares, &imoex, &rgbi, &rtsi, currencies)
		time.Sleep(5 * time.Second)
	}
}

func loadData(uri string, doc *moex.IssDocument) error {
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err = xml.Unmarshal([]byte(body), doc); err != nil {
		return (err)
	}
	return nil
}

func printFrame(shares []*share.Share, imoex *share.Share, rgbi *share.Share, rtsi *share.Share, currencies []*share.Share) {
	printHeader()
	for _, share := range shares {
		printShare(share)
	}
	printBlank()

	printShare(imoex)
	printShare(rgbi)
	printShare(rtsi)
	printBlank()

	for _, c := range currencies {
		printShare(c)
	}
	printBlank()

	printTime()
	printSeparator()
}

func printHeader() {
	fmt.Println(rowSeparator)
	fmt.Printf("| %12s %15s %17s %13s %13s %16s |\n", "SHARE", "PRICE,RUB", "PREV_PRICE,RUB", "CHANGE,RUB", "CHANGE,%", "VOLUME,RUB")
	fmt.Println(rowSeparator)
}

func printShare(s *share.Share) {
	var color string
	color = colorGreen
	if s.PriceChange() < 0 {
		color = colorRed
	}
	fmt.Printf("| %12s %s%15.1f%s %17.1f %s%13.1f %13.1f %s %15s |\n", s.Ticker, color, s.Price, colorReset, s.PrevPrice, color, s.PriceChange(), s.PriceChangePercent(), colorReset, s.FormattedVolume())
}

func printTime() {
	fmt.Printf("|         TIME  %77s |\n", time.Now().Format("2006-01-02 15:04:05"))
}

func printBlank() {
	fmt.Printf("|%93s|\n", " ")
}

func printSeparator() {
	fmt.Println(rowSeparator)
}

func populateSecurities(shares map[string]*share.Share, securities moex.IssData) error {
	if securities.Name != "securities" {
		return fmt.Errorf("securities must have name 'securities' but has '%s'", securities.Name)
	}
	for _, v := range securities.Rows {
		share, ok := shares[v.Secid]
		if ok {
			share.PrevPrice = v.PrevPrice
		}
	}
	return nil
}

func populateMarketData(shares map[string]*share.Share, marketdata moex.IssData) error {
	if marketdata.Name != "marketdata" {
		return fmt.Errorf("market data must have name 'market' but has '%s'", marketdata.Name)
	}
	for _, v := range marketdata.Rows {
		share, ok := shares[v.Secid]
		if ok {
			var p float32
			if v.Price != 0 {
				p = v.Price
			} else {
				p = v.MarketPrice2
			}
			share.Price = p
			share.Volume = v.Volume
		}
	}
	return nil
}

func getIndexData(marketdata moex.IssData, ticker string) (share.Share, error) {
	var s share.Share
	for _, v := range marketdata.Rows {
		if v.Secid == ticker {
			s = share.New(ticker)
			s.Price = v.CurrentValue
			s.PrevPrice = v.LastVlaue
			s.Volume = v.Volume
			return s, nil
		}
	}
	return s, fmt.Errorf("security %s not found", ticker)
}
