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
	log = logging.MustGetLogger("audit")

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
		log.Critical(err.Error())
		os.Exit(1)
	} else if err != nil {
		log.Error(err.Error())
	}

	// open the file
	file, err := os.Open(infile)
	if err != nil {
		log.Critical(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	log.Debugf("Opened %s", file.Name())

	// process all lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmd := parseCommand(scanner.Text())
		if err := executeCommand(cmd); err != nil {
			// if it fails, log and continue on
			log.Errorf("Command execution error! cmd # %3d message: %s", cmd.ID, err.Error())
		}
	}

	// catch read errors
	if err := scanner.Err(); err != nil {
		log.Critical(err.Error())
		os.Exit(1)
	}

	log.Debugf("Done!")
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
	log.Debugf("Parsing: %s", s)

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

	log.Debugf("Parsed as: %+v", parsed)

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
	case commands.Sell:
		status = executeSell(cmd)
	default:
		log.Warningf("Not implemented: %s", cmd.Name)
		return nil
	}

	// report our status
	if status {
		log.Debugf("Finished command %d", cmd.ID)
	} else {
		log.Debugf("Finished command %d with errors", cmd.ID)
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
		log.Errorf("Wrong number of commands: `%s`", cmd.Args)
		return false
	} else if cmd.Args[0] == "" {
		// missing
		log.Error("No amount passed to ADD")
		return false
	}

	// Convert to a centInt
	amount, err := currency.NewFromString(cmd.Args[0])
	if err != nil {
		// Bail on parse failure
		log.Error("Failed to parse currency")
		return false
	}

	// Create an account if the user needs one
	if !accountStore.HasAccount(cmd.UserID) {
		log.Noticef("Creating account for %s", cmd.UserID)
		if err := accountStore.CreateAccount(cmd.UserID); err != nil {
			log.Error(err.Error())
			return false
		}
	}

	// Add the amount
	log.Infof("Adding %s to %s", amount, cmd.UserID)
	accountStore.Accounts[cmd.UserID].AddFunds(amount)

	balance := accountStore.Accounts[cmd.UserID].Balance
	log.Infof("New balance for %s is %s", cmd.UserID, balance)

	return true
}

// Gets a quote from the quoteserver
func executeQuote(cmd command) bool {
	// Get the stock from the command
	stock := cmd.Args[0]
	if stock == "" {
		log.Error("No stock passed to QUOTE")
		return false
	}

	// get a quote for the stock. (cache will determine if a fresh one is needed)
	quote, err := quotecache.GetQuote(cmd.UserID, stock)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	log.Noticef("Got quote: %+v", quote)
	// send the quote to the user
	return true
}

func executeBuy(cmd command) bool {
	//Gotta check users money and add a reserved portion
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		log.Noticef("User %s does not have an account", account)
		return false
	}

	stockSymbol := cmd.Args[0]
	dollarAmount, err := currency.NewFromString(cmd.Args[1])

	if err != nil {
		log.Noticef("Dollar amount %s is invalid", cmd.Args[1])
		return false
	}
	//User wants to buy y worth of x shares.
	userQuote, err := quotecache.GetQuote(cmd.UserID, stockSymbol)

	if err != nil {
		log.Noticef("Quote of stock %s for user %s is invalid", stockSymbol, cmd.UserID)
		return false
	}

	wholeShares, cashRemainder := userQuote.Price.FitsInto(dollarAmount)

	if wholeShares == 0 {
		log.Notice("Amount specified to buy less than single stock unit")
		return true
	}

	log.Notice("User %s set purchase order for %d shares of stock %s", cmd.UserID, wholeShares, stockSymbol)

	dollarAmount.Sub(cashRemainder)
	account.RemoveFunds(dollarAmount)

	return account.AddToBuyQueue(stockSymbol, wholeShares, userQuote.Price)
}

func executeCommitBuy(cmd command) bool {
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		log.Infof("User %s does not have an account", cmd.UserID)
		return false
	}

	// CommitBuy has no additional args to parse! Everything is in cmd.

	// Get the most recent Buy from the user
	latestBuy, found := account.PopLatestBuy()

	// If there's no Buy or it's expired, don't change the user account
	// and log the command failure.
	if !found || latestBuy.IsExpired() {
		log.Infof("No active buys for %s", cmd.UserID)
		return false
	}

	// If there is an active Buy give the user the stock quantity.
	log.Infof("Committing buy for user %s for %d unit of %s", cmd.UserID, latestBuy.Units, latestBuy.Stock)
	log.Debugf("Before, user has %d of %s", account.GetPortfolioStockUnits(latestBuy.Stock), latestBuy.Stock)

	account.AddStockToPortfolio(latestBuy.Stock, latestBuy.Units)

	log.Debugf("After, user has %d of %s", account.GetPortfolioStockUnits(latestBuy.Stock), latestBuy.Stock)

	return true
}

func executeSell(cmd command) bool {
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		log.Noticef("User %s does not have an account", account)
		return false
	}

	stockSymbol := cmd.Args[0]
	dollarAmount, err := currency.NewFromString(cmd.Args[1])

	if err != nil {
		log.Noticef("Dollar amount %s is invalid", cmd.Args[1])
		return false
	}

	userQuote, err := quotecache.GetQuote(cmd.UserID, stockSymbol)

	if err != nil {
		log.Noticef("Quote of stock %s for user %s is invalid", stockSymbol, cmd.UserID)
		return false
	}

	wholeShares, _ := userQuote.Price.FitsInto(dollarAmount)

	if wholeShares == 0 {
		log.Notice("Amount specified to sell less than single stock unit")
		return true
	}

	log.Notice("User %s set sale order for %d shares of stock %s at %s", cmd.UserID, wholeShares, stockSymbol, userQuote.Price)

	// Do not add the money back to the account until the sale is committed

	return account.AddToSellQueue(stockSymbol, wholeShares, userQuote.Price)
}
