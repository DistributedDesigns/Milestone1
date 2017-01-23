package accounts

import (
	"errors"
	"time"

	"github.com/distributeddesigns/currency"
	"github.com/op/go-logging"
)

var (
	consoleLog = logging.MustGetLogger("console")
)

// Action : A Buy or Sell request that can expire
type Action struct {
	Time      time.Time
	Stock     string
	Units     uint
	UnitPrice currency.Currency
}

// ActionQueue : Ordered queue of actions. Oldest on left, newest on right.
type ActionQueue []Action

// Portfolio : User's stock holdings, stockName -> quantity
type Portfolio map[string]uint

// Account : State of a particular account
type Account struct {
	Balance   currency.Currency
	BuyQueue  ActionQueue
	SellQueue ActionQueue
	Portfolio Portfolio
}

// Accounts : Maps name -> Account
type Accounts map[string]*Account

// AccountStore : A collection of accouunts
type AccountStore struct {
	Accounts map[string]*Account
}

// NewAccountStore : A constructor that returns an initialized AccountStore
func NewAccountStore() *AccountStore {
	var as AccountStore
	as.Accounts = make(Accounts)
	return &as
}

// AddStockToPortfolio : Give a user some stock
func (ac *Account) AddStockToPortfolio(stock string, units uint) {
	currentUnits, ok := ac.Portfolio[stock]
	if !ok {
		currentUnits = 0
	}
	ac.Portfolio[stock] = currentUnits + units
}

// RemoveStockFromPortfolio : Remove stocks from a user
func (ac *Account) RemoveStockFromPortfolio(stock string, units uint) bool {
	currentUnits, ok := ac.Portfolio[stock]
	if !ok || currentUnits-units < 0 {
		consoleLog.Notice("User does not have enough stock to sell")
		return false
	}
	ac.Portfolio[stock] = currentUnits - units
	return true
}

// GetPortfolioStockUnits : Number of units users holds of a stock
func (ac *Account) GetPortfolioStockUnits(stock string) uint {
	return ac.Portfolio[stock]

}

// HasAccount : Checks if there's an existing account for the user
func (as AccountStore) HasAccount(name string) bool {
	_, ok := as.Accounts[name]
	return ok
}

// GetAccount ; Grab an account if it exists for the user
func (as AccountStore) GetAccount(name string) *Account {
	account, ok := as.Accounts[name]
	if !ok {
		return nil
	}
	return account
}

// AddToBuyQueue ; Add a stock S to the buy queue
func (ac *Account) AddToBuyQueue(stock string, units uint, unitPrice currency.Currency) bool {
	currentAction := Action{
		Time:      time.Now(),
		Stock:     stock,
		Units:     units,
		UnitPrice: unitPrice,
	}
	ac.BuyQueue = append(ac.BuyQueue, currentAction)
	return true
}

// AddToSellQueue ; Add a stock S to the buy queue
func (ac *Account) AddToSellQueue(stock string, units uint, unitPrice currency.Currency) bool {
	currentAction := Action{
		Time:      time.Now(),
		Stock:     stock,
		Units:     units,
		UnitPrice: unitPrice,
	}
	ac.SellQueue = append(ac.SellQueue, currentAction)
	return true
}

// CreateAccount : Initialize a new account. Fail if one already exists
func (as AccountStore) CreateAccount(name string) error {
	// Check for pre-existing accounts
	if as.HasAccount(name) {
		return errors.New("Account already exists")
	}

	// Add account with initial values
	as.Accounts[name] = &Account{}

	// Initialize the account's portfolio
	as.Accounts[name].Portfolio = make(Portfolio)

	return nil
}

// AddFunds : Increases the balance of the account
func (ac *Account) AddFunds(amount currency.Currency) {
	// Only allow > $0.00 to be added
	ac.Balance.Add(amount)
}

// RemoveFunds : Decrease balance of the account
func (ac *Account) RemoveFunds(amount currency.Currency) error {
	err := ac.Balance.Sub(amount)
	if err != nil {
		return errors.New("Insufficient Funds")
	}
	return nil
}

// PopNewest : Removes and returns the most recent action in a queue
func (aq *ActionQueue) PopNewest() (Action, bool) {
	// gobuild says:
	//   Can't use Action as nil so we send back a bool to indicate hit/miss

	// Check for empty queue
	queueLen := len(*aq)
	if queueLen == 0 {
		return Action{}, false
	}

	// Last item appended to queue will be the most recent.
	// Copy the last item then shrink the queue.
	var latestAction Action
	latestAction, *aq = (*aq)[queueLen-1], (*aq)[:queueLen-1]

	return latestAction, true
}

// IsExpired : True if the action's timestamp is older than its validity window
func (act *Action) IsExpired() bool {
	expiry := act.Time.Add(time.Second * 60)
	return time.Now().After(expiry)
}
