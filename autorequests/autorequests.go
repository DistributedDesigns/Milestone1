package autorequests

import (
	"errors"
	"github.com/distributeddesigns/currency"
)

// AutoRequest :  A buy or sell request for a user
type AutoRequest struct{ 
	Amount currency.Currency
	Trigger currency.Currency 
}

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

func (ars *AutoRequestStore) CancelAutorequest(stock, userID string) (float64, error){
	if _, found := (*ars)[stock][userID]; found {
		delete((*ars)[stock], userID)
		refundAmount := (*ars)[stock][userID].Amount.ToFloat()
		return refundAmount, nil
	} else {
		errMsg := "No request found for stock" + stock + "for user" + userID
		return 0.0, errors.New(errMsg)
	}
}

func (ars *AutoRequestStore) AutorequestExists(stock, userID string) bool {
	_, found := (*ars)[stock][userID]
	return found
}

func (ars *AutoRequestStore) GetAutorequest(stock, userID string) (AutoRequest, error) {
	userAutoRequest, found := (*ars)[stock][userID]
	if found {
		return userAutoRequest, nil
	}
	return nil, errors.New("No auto request")
}