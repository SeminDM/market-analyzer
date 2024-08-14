package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	moex "github.com/SeminDM/market/analyzer/moex"
	share "github.com/SeminDM/market/analyzer/share"
)

var shareTickers = []string{"PHOR", "SIBN", "ROSN", "SBER", "PLZL"}

var colorReset = "\033[0m"
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var rowSeparator = "---------------------------------------------------"

const issSecuritiesUri string = "https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities.xml?iss.meta=off&iss.only=marketdata,securities&marketdata.columns=SECID,LAST&securities.columns=SECID,PREVPRICE"

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
		if xml.Unmarshal([]byte(body), &doc); err != nil {
			panic(err)
		}

		shares := make(map[string]*share.Share)
		for _, ticker := range shareTickers {
			s := share.NewShare(ticker)
			shares[ticker] = &s
		}
		if err = populateSecurities(shares, doc.Data[1]); err != nil {
			panic(err)
		}
		if err = populateMarketData(shares, doc.Data[0]); err != nil {
			panic(err)
		}
		print(shares)
		time.Sleep(5 * time.Minute)
	}
}

func print(shares map[string]*share.Share) {
	for ticker, share := range shares {
		change := share.PriceChange()
		percent := share.PriceChangePercent()
		color := colorGreen
		if change < 0 {
			color = colorRed
		}

		fmt.Printf("| %5s %10.2f %12.2f %s%8.2f %8.2f%s |\n", ticker, share.Price, share.PrevPrice, color, change, percent, colorReset)
	}
	fmt.Printf("|  TIME: %40s |\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(rowSeparator)
}

func printHeader() {
	fmt.Println(rowSeparator)
	fmt.Printf("| %5s %10s %12s %8s %8s |\n", "SHARE", "PRICE", "PREV_PRICE", "CHANGE", "%")
	fmt.Println(rowSeparator)
}

func populateSecurities(shares map[string]*share.Share, securities moex.IssData) error {
	if securities.Name != "securities" {
		return fmt.Errorf("securities must have name 'securities' but has '%s'", securities.Name)
	}
	for _, v := range securities.Rows {
		share, ok := shares[v.Ticker]
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
		share, ok := shares[v.Ticker]
		if ok {
			share.Price = v.Price
		}
	}
	return nil
}
