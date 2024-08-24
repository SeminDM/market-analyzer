package share

import (
	"math"
	"strings"
)

type Share struct {
	Ticker    string
	Price     float32
	PrevPrice float32
	Volume    string
}

func New(ticker string) Share {
	return Share{Ticker: ticker, Price: -1, PrevPrice: -1}
}

func (s *Share) PriceChange() float32 {
	return s.Price - s.PrevPrice
}

func (s *Share) PriceChangePercent() float32 {
	return (s.Price - s.PrevPrice) / s.PrevPrice * 100
}

func (s *Share) FormattedVolume() string {
	v := s.Volume
	v = strings.Split(v, ".")[0]
	formatted := ""
	j := int(math.Ceil(float64(len(v)) / 3.0))

	low := len(v)
	high := -1
	for i := j; i > 0; i-- {
		high = low
		if i > 1 {
			low = high - 3
		} else {
			low = 0
		}
		formatted = s.Volume[low:high] + " " + formatted
	}
	return strings.Trim(formatted, " ")
}
