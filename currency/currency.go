package currency

import (
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

// Currency : Tracks dollars and cents as ints

var (
	log = logging.MustGetLogger("audit")
)

func Parse(dollarString string) (int64, error) {
	centString := strings.Replace(dollarString, ".", "", 1)
	centValue, err := strconv.ParseInt(centString, 10, 64)
	return centValue, err
}