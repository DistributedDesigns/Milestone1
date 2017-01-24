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

func (ars *AutoRequestStore) CancelAutorequest(stock, userID string) float64{
	if _, found := (*ars)[stock][userID]; found {
		delete((*ars)[stock], userID)
		refundAmount := (*ars)[stock][userID].Amount.ToFloat()
		return refundAmount
	} else {
		return -1.0
	}
}

func (ars *AutoRequestStore) AutorequestExists(stock, userID string, amount currency.Currency) bool{
	_, found := (*ars)[stock][userID]
	return found
}
