package main

import (
	"os"

	"github.com/op/go-logging"
)

// Globals
var (
	log = logging.MustGetLogger("audit")
)

func main() {
	initLogging()

	log.Debugf("Here's a formatted string --> %s <---", "butts")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Critical("critical")
}

func initLogging() {
	// TODO: DONE 1. Make a logger that outputs to console
	// TODO: 2. Set variable output levels based on runtime flag
	// TODO: 3. Log stuff into a file

	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)

	var consoleFormat = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{color}▶ %{level:8s}%{color:reset} %{id:03x} %{shortfile} %{message}`,
	)

	consoleBackendFormatter := logging.NewBackendFormatter(consoleBackend, consoleFormat)

	logging.SetBackend(consoleBackendFormatter)
}
