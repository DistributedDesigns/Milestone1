package autorequests

import (
	"github.com/distributeddesigns/currency"
)

// AutoRequest :  A buy or sell request for a user
type AutoRequest struct{ Amount, Trigger currency.Currency }

// AutoRequestStore : Map stock -> user -> request
type AutoRequestStore map[string](map[string]AutoRequest)

// NewAutoRequestStore :
func NewAutoRequestStore() *AutoRequestStore {
	ars := make(AutoRequestStore)

	return &ars
}

// AddBuyAmount :
func (ars *AutoRequestStore) AddAutorequest(stock, userID string, amount currency.Currency) {
	// Initialize the new user -> request map if don't find
	// any entries for the stock in the store
	if _, found := (*ars)[stock]; !found {
		(*ars)[stock] = make(map[string]AutoRequest)
	}

	// Initialize a new AutoRequest if we can't find a user.
	// This is only necessary because there's no `nil` for AutoRequest.
	if _, found := (*ars)[stock][userID]; !found {
		(*ars)[stock][userID] = AutoRequest{}
	}

	// This awkward re-assignment is here because Go doesn't let you
	// reference struct fields of indirect objects.
	// See https://github.com/golang/go/issues/3117
	request := (*ars)[stock][userID]
	request.Amount = amount
	(*ars)[stock][userID] = request
}

func (ars *AutoRequestStore) CancelAutorequest(stock, userID string) bool{
	if _, found := (*ars)[stock][userID]; found {
		delete((*ars)[stock], userID)
		return true
	} else {
		return false
	}
}

func (ars *AutoRequestStore) FindAutorequest(stock, userID string, amount currency.Currency) bool{
	if _, found := (*ars)[stock][userID]; found {
		return true
	} else {
		return false
	}
}