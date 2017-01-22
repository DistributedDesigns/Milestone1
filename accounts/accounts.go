package accounts

import (
	"errors"
	"time"

	"github.com/distributeddesigns/currency"
	"github.com/op/go-logging"

)

var (
	log = logging.MustGetLogger("audit")
)

type Action struct {
	time	time.Time
	stock	string
	units	uint
	unitPrice currency.Currency
}

// Account : State of a particular account
type Account struct {
	Balance	currency.Currency
	BuyStack, SellStack []Action
	portfolio map[string]int
}

// Accounts : Maps name -> Account
type Accounts map[string]*Account

// AccountStore : A collection of accouunts
type AccountStore struct {
	Accounts Accounts
}

// NewAccountStore : A constructor that returns an initialized AccountStore
func NewAccountStore() *AccountStore {
	var as AccountStore
	as.Accounts = make(Accounts)
	return &as
}

func (ac Account) addStockToPortfolio(stock string, units int) bool {
	currentUnits, ok := ac.portfolio[stock]
	if !ok {
		currentUnits = 0
	}
	ac.portfolio[stock] = currentUnits + units
	return true
}

func (ac Account) removeStockFromPortfolio(stock string, units int) bool {
	currentUnits, ok := ac.portfolio[stock]
	if !ok || currentUnits - units < 0{
		log.Notice("User does not have enough stock to sell")
		return false
	}
	ac.portfolio[stock] = currentUnits - units
	return true
}

func (ac Account) getPortfolioStockUnits(stock string) int {
	return ac.portfolio[stock]

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
		return nil;
	}
	return account
}

// AddToBuyStack ; Add a stock S to the buy Stack
func (ac Account) AddToBuyStack(stock string, units uint, unitPrice currency.Currency) bool {
	currentAction := Action{
		time: time.Now(),
		stock: stock,
		units: units,
		unitPrice: unitPrice,
	}
	ac.BuyStack = append(ac.BuyStack, currentAction)
	return true
}

// AddToSellStack ; Add a stock S to the buy Stack
func (ac Account) AddToSellStack(stock string, units uint, unitPrice currency.Currency) bool {
	currentAction := Action{
		time: time.Now(),
		stock: stock,
		units: units,
		unitPrice: unitPrice,
	}
	ac.SellStack = append(ac.SellStack, currentAction)
	return true
}

func (ac Account) RemoveFromBuyStack() bool
{
	if len(ac.BuyStack) <= 0
	{
		return false
	}
	else
	{
		ac.BuyStack := ac.BuyStack[:len(ac.BuyStack) - 1]
		return true
	}
}

func (ac Account) RemoveFromSellStack() bool
{
	if len(ac.SellStack) <= 0
	{
		return false
	}
	else
	{
		ac.SellStack := ac.SellStack[:len(ac.SellStack) - 1]
		return true
	}
}

// CreateAccount : Initialize a new account. Fail if one already exists
func (as AccountStore) CreateAccount(name string) error {
	// Check for pre-existing accounts
	if as.HasAccount(name) {
		return errors.New("Account already exists")
	}

	// Add account with initial values
	as.Accounts[name] = &Account{}

	return nil
}

// AddFunds : Increases the balance of the account
func (a *Account) AddFunds(amount currency.Currency) {
	// Only allow > $0.00 to be added
	a.Balance.Add(amount)
}

func (a *Account) RemoveFunds(amount currency.Currency) error {
	err := a.Balance.Sub(amount)
	if err != nil {
		return errors.New("Insufficient Funds")
	}
	return nil
}
