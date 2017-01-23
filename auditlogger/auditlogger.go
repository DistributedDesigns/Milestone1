package auditlogger

import (
	"fmt"
	"os"
	"time"

	"github.com/distributeddesigns/milestone1/commands"
	logging "github.com/op/go-logging"
)

const outdir string = "logs"

var (
	auditlogFile *os.File
	servername   string

	consoleLog = logging.MustGetLogger("console")
)

// Init : Opens a new log file and prepares attaches it to the logger.
//  	Returns a callback that will write the XML footer and close the
//		file reference.
func Init() func() {
	// Get a server name from the environment
	if os.Getenv("SERVERNAME") == "" {
		servername = "UNKNOWN"
	} else {
		servername = os.Getenv("SERVERNAME")
	}

	// Create a new file based on the current time
	now := time.Now()
	auditlogFileName := fmt.Sprintf("%s/%d%02d%02dT%02d%02d%02d.xml",
		outdir, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(),
	)

	// Create the ./logs directory, if we need to
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		consoleLog.Debugf("Creating log directory at ./%s", outdir)
		if dirErr := os.Mkdir(outdir, 0755); dirErr != nil {
			// Don't bother running if we can't generate a log file.
			consoleLog.Fatalf("Couldn't create log directory. Terminating execution.\n%s", dirErr.Error())
		}
	} else if err != nil {
		// Don't swallow the error from os.Stat
		// Bail if something weird happened.
		consoleLog.Fatalf("Couldn't create log directory. Terminating execution.\n%s", err.Error())
	}

	// Open the file for writing
	var createErr error
	auditlogFile, createErr = os.Create(auditlogFileName)
	if createErr != nil {
		consoleLog.Fatalf("Couldn't create log file. Terminating execution.\n%s", createErr.Error())
	}
	// closing the audit file is the responsiblity of the caller to Init()

	// Write our logfile header
	auditlogFile.WriteString("<?xml version=\"1.0\"?>\n<log>\n")

	// Return an anonymous function that creates a closure over the audit file.
	// If we defered File.Close() in Init() we'd close the audit file
	// as soon as Init() finished.
	return func() {
		auditlogFile.WriteString("\n</log>\n")
		auditlogFile.Close()
	}
}

// LogCommand : Writes a UserCommandType to the audit log
func LogCommand(cmd commands.Command) {
	// Parse the optional fields. Initialized to "" which is the default case.
	var usernameField, stockField, fileField, fundsField string

	// With the exception of the admin DUMPLOG, all commands will have a
	// userID. We'll assign it now and deal with the admin DUMPLOG reassignment
	// in the switch. ID and command name won't have to be reassigned so we'll
	// get those when we format the output string.
	usernameField = formatUsername(cmd.UserID)

	switch cmd.Name {
	case commands.Add:
		fundsField = formatFunds(cmd.Args[0])
	case commands.Quote:
		stockField = formatStockSymbol(cmd.Args[0])
	case commands.Buy:
		stockField = formatStockSymbol(cmd.Args[0])
		fundsField = formatFunds(cmd.Args[1])
	case commands.CommitBuy:
		// No optional args
		break
	case commands.CancelBuy:
		break
	case commands.Sell:
		stockField = formatStockSymbol(cmd.Args[0])
		fundsField = formatFunds(cmd.Args[1])
	case commands.CommitSell:
		break
	case commands.CancelSell:
		break
	case commands.SetBuyAmount:
		stockField = formatStockSymbol(cmd.Args[0])
		fundsField = formatFunds(cmd.Args[1])
	case commands.SetBuyTrigger:
		stockField = formatStockSymbol(cmd.Args[0])
		fundsField = formatFunds(cmd.Args[1])
	case commands.CancelSetBuy:
		stockField = formatStockSymbol(cmd.Args[0])
	case commands.SetSellAmount:
		stockField = formatStockSymbol(cmd.Args[0])
		fundsField = formatFunds(cmd.Args[1])
	case commands.SetSellTrigger:
		stockField = formatStockSymbol(cmd.Args[0])
		fundsField = formatFunds(cmd.Args[1])
	case commands.CancelSetSell:
		stockField = formatStockSymbol(cmd.Args[0])
	case commands.DisplaySummary:
		break
	case commands.DumpLog:
		// Only way to tell difference between user / admin command
		// is to count the number of args
		if len(cmd.Args) == 0 {
			// It's an admin dump
			// FIXME: ParseCommand interpreted the filename as the user
			fileField = formatFile(cmd.UserID)
			usernameField = ""
		} else {
			// It's a user dump
			fileField = formatFile(cmd.Args[0])
		}
	}

	timeInMillisec := time.Now().Unix() * 1000

	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>%s</command>%s%s%s%s
	</userCommand>`,
		timeInMillisec, servername, cmd.ID, cmd.Name,
		usernameField, stockField, fileField, fundsField,
	)

	auditlogFile.WriteString(xmlElement)
}

func formatUsername(name string) string {
	return fmt.Sprintf("\n\t\t<username>%s</username>", name)
}

func formatStockSymbol(stock string) string {
	return fmt.Sprintf("\n\t\t<stockSymbol>%s</stockSymbol>", stock)
}

func formatFile(file string) string {
	return fmt.Sprintf("\n\t\t<filename>%s</filename>", file)
}

func formatFunds(funds string) string {
	return fmt.Sprintf("\n\t\t<funds>%s</funds>", funds)
}

// LogQuoteServerHit : Writes a QuoteServerType to the log
func LogQuoteServerHit(s string) {
	// FIXME : s needs to be pre-formatted by the caller.
	// 		Pretty bad isolation IMO. (or is it separation of concerns?)
	auditlogFile.WriteString(s)
}
