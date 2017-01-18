package quotecache

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

// Quote : Stored response from the quoteserver
type Quote struct {
	UserID    string
	Stock     string
	Price     float64
	Timestamp time.Time
	Cryptokey string
}

// Global to store cached responses.
// Maps stock name -> Quote
var cache = make(map[string]Quote)

// GetQuote : Gets the current value of the stock, hitting the local cache if it can.
func GetQuote(userID, stock string) (Quote, error) {
	// check if stock is in cache
	//   - yes, check time
	//   - no, or expired: get new quote and save to cache

	conn, err := net.DialTimeout("tcp", getQuoteServAddress(), time.Second*10)
	if err != nil {
		return Quote{}, err
	}
	defer conn.Close()

	// Send that request!
	request := fmt.Sprintf("%s,%s", stock, userID)
	conn.Write([]byte(request))

	// listen for response
	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Println("From server:", message)

	return Quote{}, nil
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
