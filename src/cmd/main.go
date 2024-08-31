package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"sort"
	"time"

	"github.com/SeminDM/market/analyzer/internal"
	moex "github.com/SeminDM/market/analyzer/moex"
	share "github.com/SeminDM/market/analyzer/share"
)

var defaultStocks = []string{"PHOR", "SIBN", "ROSN", "SBER", "MTSS", "BELU", "MDMG"}
var defaultIndices = []string{"IMOEX", "RGBI", "RTSI"}
var defaultCurrencies = []string{"GLDRUB_TOM", "CNYRUB_TOM", "USD000UTSTOM"}
var defaultFutures = []string{"BRU4", "NGU4", "SVU4", "GLDRUBF"}

var stocksTickers internal.ShareFlags
var indicesTickers internal.ShareFlags
var futuresTickers internal.ShareFlags
var currenciesTickers internal.ShareFlags

func main() {

	period := flag.Int("period", 5, "Polling period in minutes")
	flag.Var(&stocksTickers, "stocks", "Comma-separated stocks tickers")
	flag.Var(&indicesTickers, "index", "Comma-separated index tickers")
	flag.Var(&futuresTickers, "futures", "Comma-separated futures tickersa")
	flag.Var(&currenciesTickers, "currency", "Comma-separated currency tickers")
	flag.Parse()

	if len(stocksTickers) == 0 {
		stocksTickers = defaultStocks
	}
	if len(currenciesTickers) == 0 {
		currenciesTickers = defaultCurrencies
	}
	if len(futuresTickers) == 0 {
		futuresTickers = defaultFutures
	}
	if len(indicesTickers) == 0 {
		indicesTickers = defaultIndices
	}

	for i := 0; i < 1; i++ {
		var shares []*share.Share
		var currencies []*share.Share
		var futures []*share.Share
		var imoex *share.Share
		var rtsi *share.Share
		var rgbi *share.Share

		if len(stocksTickers) > 0 {
			doc, err := moex.LoadData(moex.IssStocksUri)
			if err != nil {
				panic(err)
			}
			sharesByTicker := make(map[string]*share.Share)
			for _, ticker := range stocksTickers {
				s := share.New(ticker)
				sharesByTicker[ticker] = &s
			}
			if err := populateSecurities(sharesByTicker, doc.Data[1]); err != nil {
				panic(err)
			}
			if err := populateMarketData(sharesByTicker, doc.Data[0]); err != nil {
				panic(err)
			}
			shares = make([]*share.Share, 0, len(sharesByTicker))
			for _, v := range sharesByTicker {
				shares = append(shares, v)
			}
			sort.Sort(share.ByChange(shares))
		}

		if slices.Contains(indicesTickers, "IMOEX") || slices.Contains(indicesTickers, "RGBI") {
			doc, err := moex.LoadData(moex.IssIndexUri)
			if err != nil {
				panic(err)
			}
			if slices.Contains(indicesTickers, "IMOEX") {
				imoex, err = getIndexData(doc.Data[0], "IMOEX")
				if err != nil {
					panic(err)
				}
			}
			if slices.Contains(indicesTickers, "RGBI") {
				rgbi, err = getIndexData(doc.Data[0], "RGBI")
				if err != nil {
					panic(err)
				}
			}
		}

		if slices.Contains(indicesTickers, "RTSI") {
			doc, err := moex.LoadData(moex.IssRtsiUri)
			if err != nil {
				panic(err)
			}
			rtsi, err = getIndexData(doc.Data[0], "RTSI")
			if err != nil {
				panic(err)
			}

		}

		if len(currenciesTickers) > 0 {
			doc, err := moex.LoadData(moex.IssCurrencyUri)
			if err != nil {
				panic(err)
			}

			currencyByTicker := make(map[string]*share.Share)
			for _, ticker := range currenciesTickers {
				s := share.New(ticker)
				currencyByTicker[ticker] = &s
			}
			if err := populateSecurities(currencyByTicker, doc.Data[1]); err != nil {
				panic(err)
			}
			if err := populateMarketData(currencyByTicker, doc.Data[0]); err != nil {
				panic(err)
			}
			currencies = make([]*share.Share, 0, len(currencyByTicker))
			for _, v := range currencyByTicker {
				currencies = append(currencies, v)
			}
			sort.Sort(share.ByChange(currencies))
		}

		if len(futuresTickers) > 0 {
			doc, err := moex.LoadData(moex.IssFuturesUri)
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
			futures = make([]*share.Share, 0, len(futuresByTicker))
			for _, v := range futuresByTicker {
				futures = append(futures, v)
			}
			sort.Sort(share.ByChange(futures))
		}

		printer := internal.NewPrinter(os.Stdout)
		printer.PrintFrame(shares, imoex, rgbi, rtsi, currencies, futures)
		os.Exit(0)
		time.Sleep(time.Duration(*period) * time.Minute)
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

func getIndexData(marketdata moex.IssData, ticker string) (*share.Share, error) {
	var s share.Share
	for _, v := range marketdata.Rows {
		if v.Secid == ticker {
			s = share.New(ticker)
			s.Price = v.CurrentValue
			s.PrevPrice = v.LastVlaue
			s.Volume = v.Volume
			return &s, nil
		}
	}
	return nil, fmt.Errorf("security %s not found", ticker)
}
