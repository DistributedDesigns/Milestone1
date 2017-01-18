package accounts

import (
	"errors"
	"time"
)


type BuyAction struct {
	time	time.Time
	stock	string
}

type SellAction struct {
	time	time.Time
	stock	string
}

// Account : State of a particular account
type Account struct {
	Balance float64
	BuyQueue []BuyAction
	SellQueue []SellAction
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

// AddToBuyQueue ; Add a stock S to the buy queue
func (ac Account) AddToBuyQueue(stock string) bool {
	currentAction := BuyAction{
		time: time.Now(),
		stock: stock,
	}
	ac.BuyQueue = append(ac.BuyQueue, currentAction)
	return true
}

// AddToSellQueue ; Add a stock S to the buy queue
func (ac Account) AddToSellQueue(stock string) bool {
	currentAction := SellAction{
		time: time.Now(),
		stock: stock,
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

	return nil
}

// AddFunds : Increases the balance of the account
func (a *Account) AddFunds(amount float64) error {
	// Only allow > $0.00 to be added
	if amount <= 0 {
		return errors.New("Can only add > $0.00 to accounts")
	}

	a.Balance += amount

	return nil
}
