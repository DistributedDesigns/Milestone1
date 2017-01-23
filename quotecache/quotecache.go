package quotecache

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/distributeddesigns/currency"

	"github.com/distributeddesigns/milestone1/auditlogger"
)

// Quote : Stored response from the quoteserver
type Quote struct {
	UserID        string
	Stock         string
	Price         currency.Currency
	Timestamp     time.Time
	Cryptokey     string
	TransactionID int
}

// IsExpired : True if the quotes timestamp is older than its validity window
func (q Quote) IsExpired() bool {
	expiry := q.Timestamp.Add(time.Second * 60)
	return time.Now().After(expiry)
}

//quoteCache holds quotes for each user. e.g "AAPL": {"John": John'sQuoteInstance}
var quoteCache = make(map[string]map[string]Quote)

// GetQuote : Gets the current value of the stock, hitting the local cache if it can.
func GetQuote(userID, stock string, transactionID int) (Quote, error) {
	// check if the value is in cache

	var userQuote Quote
	userMap := quoteCache[stock]
	userQuote, found := userMap[userID]
	if found && !userQuote.IsExpired() {
		//Get it from the cache
		return userQuote, nil
	}
	//Failed to get from cache, go do it outselves.

	// get it from the quote server
	err := updateQuoteCache(userID, stock)
	if err != nil {
		return Quote{}, err
	}

	// Tag the quote with the transaction that caused the server hit.
	//
	// We have to get the new quote, modify it and write it back to the
	// cache because... go doesn't like accessing properties of indexed
	// items :p. The alternative is pass the transaction ID all the way
	// down to parseQuote() but that's way too deep.
	userQuote = quoteCache[stock][userID]
	userQuote.TransactionID = transactionID
	quoteCache[stock][userID] = userQuote

	// Write the cache hit to the audit log
	// FIXME : Should be able to pass Quote to logger.
	xmlElement := fmt.Sprintf(`
	<quoteServer>
		<timestamp>%d</timestamp>
		<server>QSRV1</server>
		<transactionNum>%d</transactionNum>
		<price>%.2f</price>
		<stockSymbol>%s</stockSymbol>
		<username>%s</username>
		<quoteServerTime>%d</quoteServerTime>
		<cryptokey>%s</cryptokey>
	</quoteServer>`,
		time.Now().Unix()*1000, transactionID, userQuote.Price.ToFloat(),
		userQuote.Stock, userQuote.UserID, userQuote.Timestamp.Unix()*1000,
		userQuote.Cryptokey,
	)

	auditlogger.LogQuoteServerHit(xmlElement)

	return userQuote, nil
}

// Refreshes the stock in the global quote cache
func updateQuoteCache(userID, stock string) error {
	conn, err := net.DialTimeout("tcp", getQuoteServAddress(), time.Second*10)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send that request!
	request := fmt.Sprintf("%s,%s", stock, userID)
	conn.Write([]byte(request))

	// listen for response
	message, err := bufio.NewReader(conn).ReadString('\n')
	// when stream is done an EOF is omitted that we should ignore
	if err != nil && err != io.EOF {
		errMessage := fmt.Sprint("Bufio reader says:", err.Error())
		return errors.New(errMessage)
	}

	// Convert the raw response to a Quote
	quote, err := parseQuote(message)
	if err != nil {
		return err
	}

	quoteCache[stock] = map[string]Quote{userID: quote}

	return nil
}

// Returns the appropriate URL & Port based on the run environment.
// Conrolled via environment flags
func getQuoteServAddress() string {
	var address string

	switch os.Getenv("ENV") {
	case "PROD":
		address = "quoteserve.seng.uvic.ca:4443"
	case "DEV":
	default:
		address = "localhost:4443"
	}

	return address
}

func parseQuote(s string) (Quote, error) {
	parts := strings.Split(s, ",")

	// Does the response have all the parts we need?
	if len(parts) != 5 {
		return Quote{}, errors.New("Incorrect number of fields returned by quoteserver")
	}

	balance, err := currency.NewFromString(parts[0])

	if err != nil {
		return Quote{}, err
	}

	// Unix time has to be converted string -> int -> Time
	unixTimeInt, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return Quote{}, err
	}

	quote := Quote{
		Price:     balance,
		Stock:     parts[1],
		UserID:    parts[2],
		Timestamp: time.Unix(unixTimeInt, 0),
		Cryptokey: parts[4],
	}

	return quote, nil
}
