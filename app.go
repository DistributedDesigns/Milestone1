package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/distributeddesigns/currency"
	"github.com/op/go-logging"

	"github.com/distributeddesigns/milestone1/accounts"
	"github.com/distributeddesigns/milestone1/commands"
	"github.com/distributeddesigns/milestone1/quotecache"
)

// Globals
var (
	consoleLog = logging.MustGetLogger("audit")

	logLevel = flag.String("loglevel", "WARNING", "CRITICAL, ERROR, WARNING,  NOTICE, INFO, DEBUG")

	accountStore = accounts.NewAccountStore()
)

// I suck at namespacing and don't want to type commands.Command over and over
type command commands.Command

func main() {
	flag.Parse()
	initLogging()

	// Find the workload file and open it
	// -  Read each line and:
	// -    parse the command
	// -    execute it
	// -    log it

	// naïvely assume this is the workload file
	infile := flag.Arg(0)
	if _, err := os.Stat(infile); os.IsNotExist(err) {
		// can't find input; bail!
		consoleLog.Critical(err.Error())
		os.Exit(1)
	} else if err != nil {
		consoleLog.Error(err.Error())
	}

	// open the file
	file, err := os.Open(infile)
	if err != nil {
		consoleLog.Critical(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	consoleLog.Debugf("Opened %s", file.Name())

	// process all lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmd := parseCommand(scanner.Text())
		if err := executeCommand(cmd); err != nil {
			// if it fails, log and continue on
			consoleLog.Errorf("Command execution error! cmd # %3d message: %s", cmd.ID, err.Error())
		}
	}

	// catch read errors
	if err := scanner.Err(); err != nil {
		consoleLog.Critical(err.Error())
		os.Exit(1)
	}

	consoleLog.Debugf("Done!")
}

func initLogging() {
	// TODO: DONE 1. Make a logger that outputs to console
	// TODO: DONE 2. Set variable output levels based on runtime flag
	// TODO: 3. Log stuff into a file

	// Create a default backend
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)

	// Add output formatting
	var consoleFormat = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{color}▶ %{level:8s}%{color:reset} %{id:03d} %{shortfile} %{message}`,
	)
	consoleBackendFormatted := logging.NewBackendFormatter(consoleBackend, consoleFormat)

	// Add leveled logging
	level, err := logging.LogLevel(*logLevel)
	if err != nil {
		fmt.Println("Bad log level. Using default.") // ERROR
	}
	consoleBackendFormattedAndLeveled := logging.AddModuleLevel(consoleBackendFormatted)
	consoleBackendFormattedAndLeveled.SetLevel(level, "")

	// Attach the backend
	logging.SetBackend(consoleBackendFormattedAndLeveled)
}

func parseCommand(s string) command {
	consoleLog.Debugf("Parsing: %s", s)

	// Convert to a proper .csv, then parse
	// change `[100] STUFF,...` -> `100,STUFF,...`
	csv := strings.Replace(s, "[", "", 1)
	csv = strings.Replace(csv, "] ", ",", 1)

	// commands have mystery spaces?!
	csv = strings.Replace(csv, " ", "", -1)

	// Break it apart
	parts := strings.Split(csv, ",")

	// Almost all commands will follow this format
	// TODO: Deal with the final "DUMPLOG,./testLOG"
	ID, _ := strconv.Atoi(parts[0])
	name, _ := commands.ToCommandType(parts[1])
	parsed := command{
		ID:     ID,
		Name:   name,
		UserID: parts[2],
		Args:   parts[3:],
	}

	consoleLog.Debugf("Parsed as: %+v", parsed)

	return parsed
}

func executeCommand(cmd command) error {
	// Each command will return true if everything is okay, false if error
	var status bool

	// Filter based on the "enum" of command names
	switch cmd.Name {
	case commands.Add:
		status = executeAdd(cmd)
	case commands.Quote:
		status = executeQuote(cmd)
	case commands.Buy:
		status = executeBuy(cmd)
	case commands.CommitBuy:
		status = executeCommitBuy(cmd)
	case commands.CancelBuy:
		status = executeCancelBuy(cmd)
	case commands.Sell:
		status = executeSell(cmd)
	case commands.CommitSell:
		status = executeCommitSell(cmd)
	case commands.CancelSell:
		status = executeCancelSell(cmd)
	default:
		consoleLog.Warningf("Not implemented: %s", cmd.Name)
		return nil
	}

	// report our status
	if status {
		consoleLog.Debugf("Finished command %d", cmd.ID)
	} else {
		consoleLog.Debugf("Finished command %d with errors", cmd.ID)
	}

	return nil
}

// Add funds to the user's account
func executeAdd(cmd command) bool {
	// Finish parsing the rest of the command.
	// ADD should have an amount passed

	// Sanitize the command
	if len(cmd.Args) != 1 {
		// too many
		consoleLog.Errorf("Wrong number of commands: `%s`", cmd.Args)
		return false
	} else if cmd.Args[0] == "" {
		// missing
		consoleLog.Error("No amount passed to ADD")
		return false
	}

	// Convert to a centInt
	amount, err := currency.NewFromString(cmd.Args[0])
	if err != nil {
		// Bail on parse failure
		consoleLog.Error("Failed to parse currency")
		return false
	}

	// Create an account if the user needs one
	if !accountStore.HasAccount(cmd.UserID) {
		consoleLog.Noticef("Creating account for %s", cmd.UserID)
		if err := accountStore.CreateAccount(cmd.UserID); err != nil {
			consoleLog.Error(err.Error())
			return false
		}
	}

	// Add the amount
	consoleLog.Infof("Adding %s to %s", amount, cmd.UserID)
	accountStore.Accounts[cmd.UserID].AddFunds(amount)

	balance := accountStore.Accounts[cmd.UserID].Balance
	consoleLog.Infof("New balance for %s is %s", cmd.UserID, balance)

	return true
}

// Gets a quote from the quoteserver
func executeQuote(cmd command) bool {
	// Get the stock from the command
	stock := cmd.Args[0]
	if stock == "" {
		consoleLog.Error("No stock passed to QUOTE")
		return false
	}

	// get a quote for the stock. (cache will determine if a fresh one is needed)
	quote, err := quotecache.GetQuote(cmd.UserID, stock)
	if err != nil {
		consoleLog.Error(err.Error())
		return false
	}

	consoleLog.Noticef("Got quote: %+v", quote)
	// send the quote to the user
	return true
}

func executeBuy(cmd command) bool {
	//Gotta check users money and add a reserved portion
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		consoleLog.Noticef("User %s does not have an account", account)
		return false
	}

	stockSymbol := cmd.Args[0]
	dollarAmount, err := currency.NewFromString(cmd.Args[1])

	if err != nil {
		consoleLog.Noticef("Dollar amount %s is invalid", cmd.Args[1])
		return false
	}
	//User wants to buy y worth of x shares.
	userQuote, err := quotecache.GetQuote(cmd.UserID, stockSymbol)

	if err != nil {
		consoleLog.Noticef("Quote of stock %s for user %s is invalid", stockSymbol, cmd.UserID)
		return false
	}

	wholeShares, cashRemainder := userQuote.Price.FitsInto(dollarAmount)

	if wholeShares == 0 {
		consoleLog.Notice("Amount specified to buy less than single stock unit")
		return true
	}

	consoleLog.Infof("User %s set purchase order for %d shares of stock %s", cmd.UserID, wholeShares, stockSymbol)

	// Remove the funds from user now to prevent double spending
	dollarAmount.Sub(cashRemainder)
	account.RemoveFunds(dollarAmount)

	return account.AddToBuyQueue(stockSymbol, wholeShares, userQuote.Price)
}

func executeCommitBuy(cmd command) bool {
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		consoleLog.Infof("User %s does not have an account", cmd.UserID)
		return false
	}

	// CommitBuy has no additional args to parse! Everything is in cmd.

	// Get the most recent Buy from the user and shrink the buy queue
	newestBuy, found := account.BuyQueue.PopNewest()

	// If there's no Buy or it's expired, don't change the user account
	// and log the command failure.
	if !found || newestBuy.IsExpired() {
		consoleLog.Infof("No active buys to commit for %s", cmd.UserID)
		return false
	}

	// If there is an active Buy give the user the stock quantity.
	consoleLog.Infof("Committing buy for user %s for %d unit of %s", cmd.UserID, newestBuy.Units, newestBuy.Stock)
	consoleLog.Debugf("Before, user has %d of %s", account.GetPortfolioStockUnits(newestBuy.Stock), newestBuy.Stock)

	account.AddStockToPortfolio(newestBuy.Stock, newestBuy.Units)

	consoleLog.Debugf("After, user has %d of %s", account.GetPortfolioStockUnits(newestBuy.Stock), newestBuy.Stock)

	return true
}

func executeCancelBuy(cmd command) bool {
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		consoleLog.Infof("User %s does not have an account", cmd.UserID)
		return false
	}

	// CommitBuy has no additional args to parse! Everything is in cmd.

	// Pop the latest Buy to get it out of the queue
	newestBuy, found := account.BuyQueue.PopNewest()
	if !found {
		consoleLog.Infof("No active buys to cancel for %s", cmd.UserID)
		return false
	}

	// Return the reserve amount to the user's balance
	var reserve currency.Currency
	reserve.Add(newestBuy.UnitPrice)
	reserve.Mul(float64(newestBuy.Units))

	consoleLog.Infof("Cancel buy for %s. Adding back %s", newestBuy.Stock, reserve)
	consoleLog.Debugf("Before, user balance %s", account.Balance)

	account.AddFunds(reserve)

	consoleLog.Debugf("After, user balance %s", account.Balance)

	return true
}

func executeSell(cmd command) bool {
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		consoleLog.Noticef("User %s does not have an account", account)
		return false
	}

	stockSymbol := cmd.Args[0]
	dollarAmount, err := currency.NewFromString(cmd.Args[1])

	if err != nil {
		consoleLog.Noticef("Dollar amount %s is invalid", cmd.Args[1])
		return false
	}

	userQuote, err := quotecache.GetQuote(cmd.UserID, stockSymbol)

	if err != nil {
		consoleLog.Noticef("Quote of stock %s for user %s is invalid", stockSymbol, cmd.UserID)
		return false
	}

	wholeShares, _ := userQuote.Price.FitsInto(dollarAmount)

	if wholeShares == 0 {
		consoleLog.Notice("Amount specified to sell less than single stock unit")
		return true
	}

	consoleLog.Infof("User %s set sale order for %d shares of stock %s at %s", cmd.UserID, wholeShares, stockSymbol, userQuote.Price)

	// Remove stock now to prevent double selling
	if stockWasRemoved := account.RemoveStockFromPortfolio(stockSymbol, wholeShares); !stockWasRemoved {
		return false
	}

	// Make the new sell order and report success
	return account.AddToSellQueue(stockSymbol, wholeShares, userQuote.Price)
}

func executeCommitSell(cmd command) bool {
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		consoleLog.Infof("User %s does not have an account", cmd.UserID)
		return false
	}

	// CommitSell has no additional args to parse! Everything is in cmd.

	// Pop the latest Sell to get it out of the queue
	newestSell, found := account.SellQueue.PopNewest()
	if !found {
		consoleLog.Infof("No active sells to cancel for %s", cmd.UserID)
		return false
	}

	// Add the profit of the sale to the user's account
	var profit currency.Currency
	profit.Add(newestSell.UnitPrice)
	profit.Mul(float64(newestSell.Units))

	consoleLog.Infof("Commit sell for %s of %d units of %s at %s. Adding %s",
		cmd.UserID, newestSell.Units, newestSell.Stock, newestSell.UnitPrice, profit,
	)
	consoleLog.Debugf("Before, user balance %s", account.Balance)

	account.AddFunds(profit)

	consoleLog.Debugf("After, user balance %s", account.Balance)

	return true
}

func executeCancelSell(cmd command) bool {
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		consoleLog.Infof("User %s does not have an account", cmd.UserID)
		return false
	}

	// CancelSell has no additional args to parse! Everything is in cmd.

	newestSell, found := account.SellQueue.PopNewest()
	if !found {
		consoleLog.Infof("No active sells to cancel for %s", cmd.UserID)
		return false
	}

	// Add the stock back to the user's portfolio
	consoleLog.Infof("Cancel sell for %s. Adding back %d units", newestSell.Stock, newestSell.Units)
	consoleLog.Debugf("Before, user portfolio: %d x %s", account.Portfolio[newestSell.Stock], newestSell.Stock)

	account.AddStockToPortfolio(newestSell.Stock, newestSell.Units)

	consoleLog.Debugf("After, user portfolio: %d x %s", account.Portfolio[newestSell.Stock], newestSell.Stock)

	return true
}
