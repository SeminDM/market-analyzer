package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/SeminDM/market/analyzer/internal"
	moex "github.com/SeminDM/market/analyzer/moex"
	share "github.com/SeminDM/market/analyzer/share"
)

var shareTickers = []string{"PHOR", "SIBN", "ROSN", "SBER", "PLZL", "BELU"}
var currencyTickers = []string{"GLDRUB_TOM", "CNYRUB_TOM", "USD000UTSTOM"}
var futuresTickers = []string{"BRU4", "NGU4", "SVU4", "GLDRUBF"}

func main() {
	for i := 0; i < 1000; i++ {
		doc, err := moex.LoadData(moex.IssStocksUri)
		if err != nil {
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

		doc, err = moex.LoadData(moex.IssIndexUri)
		if err != nil {
			panic(err)
		}
		imoex, err := getIndexData(doc.Data[0], "IMOEX")
		if err != nil {
			panic(err)
		}
		rgbi, err := getIndexData(doc.Data[0], "RGBI")
		if err != nil {
			panic(err)
		}

		doc, err = moex.LoadData(moex.IssRtsiUri)
		if err != nil {
			panic(err)
		}
		rtsi, err := getIndexData(doc.Data[0], "RTSI")
		if err != nil {
			panic(err)
		}

		doc, err = moex.LoadData(moex.IssCurrencyUri)
		if err != nil {
			panic(err)
		}

		currencyByTicker := make(map[string]*share.Share)
		for _, ticker := range currencyTickers {
			s := share.New(ticker)
			currencyByTicker[ticker] = &s
		}
		if err := populateSecurities(currencyByTicker, doc.Data[1]); err != nil {
			panic(err)
		}
		if err := populateMarketData(currencyByTicker, doc.Data[0]); err != nil {
			panic(err)
		}
		currencies := make([]*share.Share, 0, len(currencyByTicker))
		for _, v := range currencyByTicker {
			currencies = append(currencies, v)
		}
		sort.Sort(share.ByChange(currencies))

		doc, err = moex.LoadData(moex.IssFuturesUri)
		if err != nil {
			panic(err)
		}

		futuresByTicker := make(map[string]*share.Share)
		for _, ticker := range futuresTickers {
			s := share.New(ticker)
			futuresByTicker[ticker] = &s
		}
		if err := populateSecurities(futuresByTicker, doc.Data[1]); err != nil {
			panic(err)
		}
		if err := populateMarketData(futuresByTicker, doc.Data[0]); err != nil {
			panic(err)
		}
		futures := make([]*share.Share, 0, len(futuresByTicker))
		for _, v := range futuresByTicker {
			futures = append(futures, v)
		}
		sort.Sort(share.ByChange(futures))

		printer := internal.NewPrinter(os.Stdout)
		printer.PrintFrame(shares, &imoex, &rgbi, &rtsi, currencies, futures)
		time.Sleep(5 * time.Second)
	}
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
