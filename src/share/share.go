package share

type Share struct {
	Ticker    string
	Price     float32
	PrevPrice float32
}

func NewShare(ticker string) Share {
	return Share{Ticker: ticker, Price: -1, PrevPrice: -1}
}

func (s *Share) PriceChange() float32 {
	return s.PrevPrice - s.Price
}

func (s *Share) PriceChangePercent() float32 {
	return (s.PrevPrice - s.Price) / s.PrevPrice
}
