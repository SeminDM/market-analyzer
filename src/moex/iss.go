package moex

import (
	"encoding/xml"
	"io"
	"net/http"
)

const IssStocksUri string = "https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities.xml?iss.meta=off&iss.only=marketdata,securities&marketdata.columns=SECID,LAST,VALTODAY&securities.columns=SECID,PREVPRICE"
const IssIndexUri string = "https://iss.moex.com/iss/engines/stock/markets/index/boards/SNDX/securities.xml?iss.meta=off&iss.only=marketdata&marketdata.columns=SECID,LASTVALUE,CURRENTVALUE,VALTODAY&securities=IMOEX,RGBI"
const IssRtsiUri string = "https://iss.moex.com/iss/engines/stock/markets/index/boards/RTSI/securities.xml?iss.meta=off&iss.only=marketdata&marketdata.columns=SECID,LASTVALUE,CURRENTVALUE,VALTODAY"
const IssCurrencyUri string = "https://iss.moex.com/iss/engines/currency/markets/selt/securities.xml?iss.meta=off&iss.only=marketdata,securities&securities=CETS:USD000UTSTOM,CETS:GLDRUB_TOM,CETS:CNYRUB_TOM"
const IssFuturesUri string = "https://iss.moex.com/iss/engines/futures/markets/forts/boards/RFUD/securities.xml?iss.meta=off&iss.only=marketdata,securities"

type IssDocument struct {
	Data []IssData `xml:"data"`
}

type IssData struct {
	Name string   `xml:"id,attr"`
	Rows []IssRow `xml:"rows>row"`
}

type IssRow struct {
	XMLName      xml.Name `xml:"row"`
	Secid        string   `xml:"SECID,attr"`
	Price        float32  `xml:"LAST,attr"`
	PrevPrice    float32  `xml:"PREVPRICE,attr"`
	Volume       string   `xml:"VALTODAY,attr"`
	LastVlaue    float32  `xml:"LASTVALUE,attr"`
	CurrentValue float32  `xml:"CURRENTVALUE,attr"`
	MarketPrice2 float32  `xml:"MARKETPRICE2,attr"`
}

func LoadData(uri string) (*IssDocument, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var doc IssDocument
	if err = xml.Unmarshal([]byte(body), &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
