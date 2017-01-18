package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/op/go-logging"

	"./accounts"
	"./commands"
	"./quotecache"
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
		break
	case commands.Quote:
		status = executeQuote(cmd)
		break
	case commands.Buy:
		status = executeBuy(cmd)
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

	// Convert to a float
	amount, err := strconv.ParseFloat(cmd.Args[0], 64)
	if err != nil {
		// Bail on parse failure
		log.Error(err.Error())
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
	log.Infof("Adding %.2f to %s", amount, cmd.UserID)
	if err := accountStore.Accounts[cmd.UserID].AddFunds(amount); err != nil {
		log.Error(err.Error())
		return false
	}
	balance := accountStore.Accounts[cmd.UserID].Balance
	log.Infof("New balance for %s is %.2f", cmd.UserID, balance)

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
	}

	return account.AddToBuyQueue(cmd.Args[0])
}

func executeSell(cmd command) bool {
	account := accountStore.GetAccount(cmd.UserID)

	if account == nil {
		log.Noticef("User %s does not have an account", account)
	}

	return account.AddToSellQueue(cmd.Args[0])
}


