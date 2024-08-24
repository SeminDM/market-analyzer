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

var colorReset = "\033[0m"
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var rowSeparator = "--------------------------------------------------------------------"

const issSecuritiesUri string = "https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities.xml?iss.meta=off&iss.only=marketdata,securities&marketdata.columns=SECID,LAST,VALTODAY&securities.columns=SECID,PREVPRICE"
const issIndexUri string = "https://iss.moex.com/iss/engines/stock/markets/index/boards/SNDX/securities.xml?iss.meta=off&iss.only=marketdata&marketdata.columns=SECID,LASTVALUE,CURRENTVALUE,VALTODAY"

func main() {
	printHeader()
	for i := 0; i < 1000; i++ {
		resp, err := http.Get(issSecuritiesUri)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		var doc moex.IssDocument
		if err = xml.Unmarshal([]byte(body), &doc); err != nil {
			panic(err)
		}

		sharesByTicker := make(map[string]*share.Share)
		for _, ticker := range shareTickers {
			s := share.New(ticker)
			sharesByTicker[ticker] = &s
		}
		if err = populateSecurities(sharesByTicker, doc.Data[1]); err != nil {
			panic(err)
		}
		if err = populateMarketData(sharesByTicker, doc.Data[0]); err != nil {
			panic(err)
		}
		shares := make([]*share.Share, 0, len(sharesByTicker))
		for _, v := range sharesByTicker {
			shares = append(shares, v)
		}
		sort.Sort(share.ByChange(shares))

		resp2, err := http.Get(issIndexUri)
		if err != nil {
			panic(err)
		}
		body2, err := io.ReadAll(resp2.Body)
		if err != nil {
			panic(err)
		}
		defer resp2.Body.Close()
		var doc2 moex.IssDocument
		if err = xml.Unmarshal([]byte(body2), &doc2); err != nil {
			panic(err)
		}
		imoex, err := getIMoexData(doc2)
		if err != nil {
			panic(err)
		}

		print(shares, &imoex)
		time.Sleep(5 * time.Second)
	}
}

func print(shares []*share.Share, imoex *share.Share) {
	for _, share := range shares {
		change := share.PriceChange()
		percent := share.PriceChangePercent()
		color := colorGreen
		if change < 0 {
			color = colorRed
		}

		fmt.Printf("| %5s %s%10.1f%s %12.1f %s%8.1f %8.1f %s %15s |\n", share.Ticker, color, share.Price, colorReset, share.PrevPrice, color, change, percent, colorReset, share.FormattedVolume())
	}
	fmt.Printf("|%66s|\n", " ")

	change := imoex.PriceChange()
	percent := imoex.PriceChangePercent()
	color := colorGreen
	if change < 0 {
		color = colorRed
	}
	fmt.Printf("| %5s %s%10.1f%s %12.1f %s%8.1f %8.1f %s %15s |\n", imoex.Ticker, color, imoex.Price, colorReset, imoex.PrevPrice, color, change, percent, colorReset, imoex.FormattedVolume())
	fmt.Printf("|%66s|\n", " ")

	fmt.Printf("|  TIME: %57s |\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(rowSeparator)
}

func printHeader() {
	fmt.Println(rowSeparator)
	fmt.Printf("| %5s %10s %12s %8s %8s %16s |\n", "SHARE", "PRICE", "PREV_PRICE", "CHANGE", "%", "VOLUME")
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
			share.Price = v.Price
			share.Volume = v.Volume
		}
	}
	return nil
}

func getIMoexData(marketdata moex.IssDocument) (share.Share, error) {
	var s share.Share
	for _, v := range marketdata.Data[0].Rows {
		if v.Secid == "IMOEX" {
			s = share.New("IMOEX")
			s.Price = v.CurrentValue
			s.PrevPrice = v.LastVlaue
			s.Volume = v.Volume
			return s, nil
		}
	}
	return s, fmt.Errorf("security IMOEX not found")
}
