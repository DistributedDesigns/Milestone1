package commands

import (
	"errors"
	"strings"
)

// CommandType : All valid command names
type CommandType int

// Command : Maps to parts of workload file commands
type Command struct {
	ID     int
	Name   CommandType
	UserID string
	Args   []string
}

// Command enum!
const (
	Add CommandType = iota
	Quote
	Buy
	CommitBuy
	CancelBuy
	Sell
	CommitSell
	CancelSell
	SetBuyAmount
	CancelSetBuy
	SetBuyTrigger
	SetSellAmount
	SetSellTrigger
	CancelSetSell
	DisplaySummary
	DumpLog
)

var commandNames = []string{
	"ADD",
	"QUOTE",
	"BUY",
	"COMMIT_BUY",
	"CANCEL_BUY",
	"SELL",
	"COMMIT_SELL",
	"CANCEL_SELL",
	"SET_BUY_AMOUNT",
	"CANCEL_SET_BUY",
	"SET_BUY_TRIGGER",
	"SET_SELL_AMOUNT",
	"SET_SELL_TRIGGER",
	"CANCEL_SET_SELL",
	"DISPLAY_SUMMARY",
	"DUMPLOG",
}

// String representation of the Command enum
func (c CommandType) String() string {
	return commandNames[c]
}

// ToCommandType : Convert string -> CommandType enum
func ToCommandType(cmd string) (CommandType, error) {
	for i, name := range commandNames {
		if strings.EqualFold(name, cmd) {
			return CommandType(i), nil
		}
	}

	return CommandType(0), errors.New("Not a valid command type")
}
