package main

import (
	"bufio"
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

// Globals
var (
	log = logging.MustGetLogger("audit")

	logLevel = flag.Int("loglevel", 4, "CRITICAL: 0, ERROR: 1, WARNING: 2, NOTICE: 3, INFO: 4, DEBUG: 5")
)

// All valid command names
type commandType int

// Command enum!
const (
	Add commandType = iota
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

// String representation of the commandType enum
func (c commandType) String() string {
	return commandNames[c]
}

// Convert string -> commandType enum
func toCommandType(cmd string) (commandType, error) {
	for i, name := range commandNames {
		if strings.EqualFold(name, cmd) {
			return commandType(i), nil
		}
	}

	return commandType(0), errors.New("Not a valid command type")
}

type command struct {
	ID     int
	name   commandType
	userID string
	args   []string
}

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

	log.Infof("Opened %s", file.Name())

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

	log.Info("Done!")
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
	consoleBackendFormattedAndLeveled := logging.AddModuleLevel(consoleBackendFormatted)
	consoleBackendFormattedAndLeveled.SetLevel(logging.Level(*logLevel), "")

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
	name, _ := toCommandType(parts[1])
	parsed := command{
		ID:     ID,
		name:   name,
		userID: parts[2],
		args:   parts[3:],
	}

	log.Debugf("Parsed as: %+v", parsed)

	return parsed
}

func executeCommand(cmd command) error {
	switch cmd.name {
	case Add:
		// do ADD
		break
	default:
		log.Noticef("Not implemented: %s", cmd.name)
		return nil
	}

	log.Infof("Finished command %d", cmd.ID)

	return nil
}
