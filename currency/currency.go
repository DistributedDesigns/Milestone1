package currency

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

// Currency : Tracks dollars and cents as ints

var (
	log = logging.MustGetLogger("audit")
)

type Currency struct {
	Dollars int64
	Cents int64
}

func ParseFloatToCurrency(floatString string) (Currency, bool) {
	dollarCentString := strings.Split(floatString, ".")
	parsedDollars, errDollar := strconv.ParseInt(dollarCentString[0], 10, 32)
	parsedCents, errCents := strconv.ParseInt(dollarCentString[1], 10, 32)
	if errDollar != nil || errCents != nil {
		return Currency{}, false
	}
	return Currency{
		Dollars: parsedDollars,
		Cents: parsedCents,
	}, true
}

func (curr Currency) Add(c1 Currency) {
	curr.Dollars += c1.Dollars + int64((curr.Cents + c1.Cents) /  100)
	curr.Cents += (curr.Cents + c1.Cents) % 100
}

func (curr Currency) GetWholeShares(stockPrice Currency) int {
	top, errTop := strconv.ParseFloat(fmt.Sprintf("%f.%f", curr.Dollars, curr.Cents), 64)
	bottom, errBottom := strconv.ParseFloat(fmt.Sprintf("%f.%f", curr.Dollars, curr.Cents), 64)
	if errTop == nil || errBottom == nil {
		log.Notice("Failed Currency conversion")
	}

	return int(top/bottom)
}