package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

// Globals
var (
	log = logging.MustGetLogger("audit")
)

type command struct {
	ID     int
	name   string
	userID string
	args   []string
}

func main() {
	initLogging()

	// Find the workload file and open it
	// -  Read each line and:
	// -    parse the command
	// -    execute it
	// -    log it

	// naïvely assume this is the workload file
	infile := os.Args[1]
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
}

func initLogging() {
	// TODO: DONE 1. Make a logger that outputs to console
	// TODO: 2. Set variable output levels based on runtime flag
	// TODO: 3. Log stuff into a file

	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)

	var consoleFormat = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{color}▶ %{level:8s}%{color:reset} %{id:03d} %{shortfile} %{message}`,
	)

	consoleBackendFormatter := logging.NewBackendFormatter(consoleBackend, consoleFormat)

	logging.SetBackend(consoleBackendFormatter)
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
	parsed := command{
		ID:     ID,
		name:   parts[1],
		userID: parts[2],
		args:   parts[3:],
	}

	log.Debugf("Parsed as: %+v", parsed)

	return parsed
}

func executeCommand(cmd command) error {
	return nil
}
