package moex

import (
	"encoding/xml"
)

type IssDocument struct {
	Data []IssData `xml:"data"`
}

type IssData struct {
	Name string   `xml:"id,attr"`
	Rows []IssRow `xml:"rows>row"`
}

type IssRow struct {
	XMLName   xml.Name `xml:"row"`
	Ticker    string   `xml:"SECID,attr"`
	Price     float32  `xml:"LAST,attr"`
	PrevPrice float32  `xml:"PREVPRICE,attr"`
}
