package main

import (
	// Standard
	"flag"
	"fmt"
	"os"
	"path/filepath"

	// 3rd Party
	"github.com/fatih/color"

	// reaper
	"github.com/infosechoudini/reaper/pkg"
	//"github.com/infosechoudini/reaper/pkg/banner"
	"github.com/infosechoudini/reaper/pkg/cli"
	"github.com/infosechoudini/reaper/pkg/core"
	"github.com/infosechoudini/reaper/pkg/logging"
	"github.com/infosechoudini/reaper/pkg/servers/http2"
)

// Global Variables
var build = "nonRelease"
var psk = "reaper"

func main() {
	logging.Server("Starting reaper Server version " + reaper.Version + " build " + reaper.Build)

	flag.BoolVar(&core.Verbose, "v", false, "Enable verbose output")
	flag.BoolVar(&core.Debug, "debug", false, "Enable debug output")
	port := flag.Int("p", 443, "reaper Server Port")
	ip := flag.String("i", "127.0.0.1", "The IP address of the interface to bind to")
	proto := flag.String("proto", "h2", "Protocol for the agent to connect with [h2, hq]")
	crt := flag.String("x509cert", filepath.Join(string(core.CurrentDir), "data", "x509", "server.crt"),
		"The x509 certificate for the HTTPS listener")
	key := flag.String("x509key", filepath.Join(string(core.CurrentDir), "data", "x509", "server.key"),
		"The x509 certificate key for the HTTPS listener")
	flag.StringVar(&psk, "psk", psk, "Pre-Shared Key used to encrypt initial communications")
	flag.Usage = func() {
		color.Blue("#################################################")
		color.Blue("#\t\treaper SERVER\t\t\t#")
		color.Blue("#################################################")
		color.Blue("Version: " + reaper.Version)
		color.Blue("Build: " + build)
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()

	color.Blue("REAPER")
	color.Blue("\t\t   Version: %s", reaper.Version)
	color.Blue("\t\t   Build: %s", build)

	// Start reaper Command Line Interface
	go cli.Shell()

	// Start reaper Server to listen for agents
	server, err := http2.New(*ip, *port, *proto, *key, *crt, psk)
	if err != nil {
		color.Red(fmt.Sprintf("[!]There was an error creating a new server instance:\r\n%s", err.Error()))
		os.Exit(1)
	} else {
		err := server.Run()
		if err != nil {
			color.Red(fmt.Sprintf("[!]There was an error starting the server:\r\n%s", err.Error()))
			os.Exit(1)
		}
	}
}

// TODO add CSRF tokens
// TODO check if agentLog exists even outside of InitialCheckIn
// TODO readline for file paths to use with upload
// TODO handle file names containing a space for upload/download
