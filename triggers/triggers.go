package triggers

import (
	"github.com/distributeddesigns/currency"
)

// AutoRequest :  A buy or sell request for a user
type Trigger struct{
	Amount currency.Currency
	Trigger currency.Currency
}

// AutoRequestStore : Map stock -> user -> request
type TriggerStore map[string](map[string]Trigger)

// NewAutoRequestStore :
func NewTriggerStore() *TriggerStore {
	ars := make(TriggerStore)

	return &ars
}

func (ts TriggerStore) CreateNewTrigger(stock string, userID string, Amount currency.Currency) bool {
	return true
}