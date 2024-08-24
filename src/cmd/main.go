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
var rowSeparator = "----------------------------------------------------------------------------------------"

const issSecuritiesUri string = "https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities.xml?iss.meta=off&iss.only=marketdata,securities&marketdata.columns=SECID,LAST,VALTODAY&securities.columns=SECID,PREVPRICE"
const issIndexUri string = "https://iss.moex.com/iss/engines/stock/markets/index/boards/SNDX/securities.xml?iss.meta=off&iss.only=marketdata&marketdata.columns=SECID,LASTVALUE,CURRENTVALUE,VALTODAY"
const issRtsiUri string = "https://iss.moex.com/iss/engines/stock/markets/index/boards/RTSI/securities.xml?iss.meta=off&iss.only=marketdata&marketdata.columns=SECID,LASTVALUE,CURRENTVALUE,VALTODAY"

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

		if err := loadData(issRtsiUri, &doc); err != nil {
			panic(err)
		}
		rtsi, err := getIndexData(doc.Data[3], "RTSI")
		if err != nil {
			panic(err)
		}

		printFrame(shares, &imoex, &rtsi)
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

func printFrame(shares []*share.Share, imoex *share.Share, rtsi *share.Share) {
	printHeader()
	for _, share := range shares {
		printShare(share)
	}
	printBlank()

	printShare(imoex)
	printShare(rtsi)
	printBlank()

	printTime()
	printSeparator()
}

func printHeader() {
	fmt.Println(rowSeparator)
	fmt.Printf("| %5s %15s %17s %13s %13s %16s |\n", "SHARE", "PRICE,RUB", "PREV_PRICE,RUB", "CHANGE,RUB", "CHANGE,%", "VOLUME,RUB")
	fmt.Println(rowSeparator)
}

func printShare(s *share.Share) {
	var color string
	color = colorGreen
	if s.PriceChange() < 0 {
		color = colorRed
	}
	fmt.Printf("| %5s %s%15.1f%s %17.1f %s%13.1f %13.1f %s %15s |\n", s.Ticker, color, s.Price, colorReset, s.PrevPrice, color, s.PriceChange(), s.PriceChangePercent(), colorReset, s.FormattedVolume())
}

func printTime() {
	fmt.Printf("|  TIME  %77s |\n", time.Now().Format("2006-01-02 15:04:05"))
}

func printBlank() {
	fmt.Printf("|%86s|\n", " ")
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
			share.Price = v.Price
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
