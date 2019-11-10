package logging

import (
	// Standard
	"fmt"
	"os"
	"path/filepath"
	"time"

	// 3rd Party
	"github.com/fatih/color"

	// reaper
	"github.com/infosechoudini/reaper/pkg/core"
)

var serverLog *os.File

func init() {

	// Server Logging
	if _, err := os.Stat(filepath.Join(core.CurrentDir, "data", "log", "reaperServerLog.txt")); os.IsNotExist(err) {
		errM := os.MkdirAll(filepath.Join(core.CurrentDir, "data", "log"), 0750)
		if errM != nil {
			message("warn", "there was an error creating the log directory")
		}
		serverLog, errC := os.Create(filepath.Join(core.CurrentDir, "data", "log", "reaperServerLog.txt"))
		if errC != nil {
			message("warn", "there was an error creating the reaperServerLog.txt file")
			return
		}
		// Change the file's permissions
		errChmod := serverLog.Chmod(0640)
		if errChmod != nil {
			message("warn", fmt.Sprintf("there was an error changing the file permissions for the agent log:\r\n%s", errChmod.Error()))
		}
		if core.Debug {
			message("debug", fmt.Sprintf("Created server log file at: %s\\data\\log\\reaperServerLog.txt", core.CurrentDir))
		}
	}

	var errLog error
	serverLog, errLog = os.OpenFile(filepath.Join(core.CurrentDir, "data", "log", "reaperServerLog.txt"), os.O_APPEND|os.O_WRONLY, 0600)
	if errLog != nil {
		message("warn", "there was an error with the reaper Server log file")
		message("warn", errLog.Error())
	}
}

// Server writes a log entry into the server's log file
func Server(logMessage string) {
	_, err := serverLog.WriteString(fmt.Sprintf("[%s]%s\r\n", time.Now().UTC().Format(time.RFC3339), logMessage))
	if err != nil {
		message("warn", "there was an error writing to the reaper Server log file")
	}
}

// Message is used to print a message to the command line
func message(level string, message string) {
	switch level {
	case "info":
		color.Cyan("[i]" + message)
	case "note":
		color.Yellow("[-]" + message)
	case "warn":
		color.Red("[!]" + message)
	case "debug":
		color.Red("[DEBUG]" + message)
	case "success":
		color.Green("[+]" + message)
	default:
		color.Red("[_-_]Invalid message level: " + message)
	}
}

// TODO configure all message to be displayed on the CLI to be returned as errors and not written to the CLI here
