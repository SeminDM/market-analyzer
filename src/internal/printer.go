package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/SeminDM/market/analyzer/share"
)

const ColorReset = "\033[0m"
const ColorRed = "\033[31m"
const ColorGreen = "\033[32m"
const DashSeparator = "----------------------------------------------------------------------------------------------\n"

type Printer struct {
	Destination *os.File
}

func NewPrinter(dest *os.File) Printer {
	p := Printer{Destination: dest}
	return p
}

func (p *Printer) PrintFrame(shares []*share.Share, imoex *share.Share, rgbi *share.Share, rtsi *share.Share, currencies []*share.Share, futures []*share.Share) {
	p.printHeader()
	for _, share := range shares {
		p.printShare(share)
	}
	p.printBlank()

	p.printShare(imoex)
	p.printShare(rgbi)
	p.printShare(rtsi)
	p.printBlank()

	for _, c := range currencies {
		p.printShare(c)
	}
	p.printBlank()

	for _, c := range futures {
		p.printShare(c)
	}
	p.printBlank()

	p.printTime()
	p.printSeparator()
}

func (p *Printer) printShare(s *share.Share) {
	var color string
	color = ColorGreen
	if s.PriceChange() < 0 {
		color = ColorRed
	}
	p.PrintString(fmt.Sprintf("| %12s %s%15.1f%s %17.1f %s%13.1f %13.1f %s %15s |\n", s.Ticker, color, s.Price, ColorReset, s.PrevPrice, color, s.PriceChange(), s.PriceChangePercent(), ColorReset, s.FormattedVolume()))
}

func (p *Printer) printHeader() {
	p.printSeparator()
	p.PrintString(fmt.Sprintf("| %12s %15s %17s %13s %13s %16s |\n", "SHARE", "PRICE,RUB", "PREV_PRICE,RUB", "CHANGE,RUB", "CHANGE,%", "VOLUME,RUB"))
	p.printSeparator()
}

func (p *Printer) printTime() {
	p.PrintString(fmt.Sprintf("|         TIME  %77s |\n", time.Now().Format("2006-01-02 15:04:05")))
}

func (p *Printer) printBlank() {
	p.PrintString(fmt.Sprintf("|%93s|\n", " "))
}

func (p *Printer) printSeparator() {
	p.PrintString(DashSeparator)
}

func (p *Printer) PrintString(s string) {
	n, err := p.Destination.WriteString(s)
	if err != nil {
		panic(err)
	}
	if n != len(s) {
		panic(fmt.Errorf("%d bytes written instead of %d for string %s", n, len(s), s))
	}
}
