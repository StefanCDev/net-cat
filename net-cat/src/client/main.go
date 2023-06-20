package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// DEFINE program constants
const (
	busSize      int    = 1024
	expectedArgc int    = 3
	portMin      uint16 = 1000
	portMax      uint16 = 65535
	protocol     string = "tcp4"
)

// END DEFINITION

// DEFINE MessageBus struct and methods
type MessageBus struct {
	Queue chan []byte
}

func (m MessageBus) Enqueue(buffer []byte) {
	m.Queue <- buffer
}

func (m MessageBus) Dequeue() []byte {
	return <-m.Queue
}

// END DEFINITION

// DEFINE program global variables
var msgBus MessageBus
var inputBuffer []byte = make([]byte, busSize)
var connected bool = false
var name string
var screenMutex sync.Mutex
var history []byte

// END DEFINITION

func main() {
	// if runtime.GOOS == "windows" {
	// 	fmt.Println("We don't respect that round here.")
	// 	os.Exit(1)
	// }

	// Log usage message if incorrect number of command-line arguments
	if len(os.Args) != expectedArgc {
		fmt.Println("[USAGE]: ./nc $IP $Port")
		os.Exit(1)
	}
	// Connect to server
	connection := connectToServer()
	// Instantiate message bus
	msgBus = MessageBus{Queue: make(chan []byte, busSize)}
	// Print greeting and prompt user to set name
	printGreeting()
	name = setName()
	// Run two go routines responsible for periodically requesting chat history
	// and for handling communication to/from the server
	go historyRequestHandler()
	go communicationHandler(connection)
	disableBufferingAndEcho()
	run()
}

// DEFINE go routine functions

// Designed to run in a separate go routine,
// this function periodically queues a chat history
// request to be sent to the server as long as the client is connected.
func historyRequestHandler() {
	for connected {
		msgBus.Enqueue([]byte("hist\n"))
		time.Sleep(time.Second * 1)
	}
}

// Designed to run in a separate goroutine,
// this function continually dequeues the first message from
// the message bus, and sends it to the server. It also handles
// refreshing the display whenever the chat history is updated.
// It is also responsible for closing the client socket connection.
func communicationHandler(connection net.Conn) {
	br := bufio.NewReader(connection)
	for connected {
		buffer := msgBus.Dequeue()
		_, err := connection.Write(buffer)
		if err != nil {
			log.Fatal(err.Error())
		}
		msgType := string(buffer[0:4])
		if msgType == "quit" {
			connected = false
			break
		}

		res, err := br.ReadBytes(0)
		if len(res) > 1 && len(res) != len(history)+1 {
			res = res[0 : len(res)-1]
			history = res
			updateDisplay(true)
		}
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	connection.Close()
	fmt.Printf("\nConnection closed.\n")
	enableBufferingAndEcho()
	os.Exit(0)
}

func run() {
	// Get user input as long as the client remains connected to server
	for connected {
		// By default, message type is "text"
		msgType := []byte("text")
		// Create a buffer to capture user input as it is typed
		inputBuffer = make([]byte, busSize)
		// Read one byte at a time from standard input (terminal)
		for i := 0; i < busSize; i++ {
			b := make([]byte, 1)
			os.Stdin.Read(b)
			// If a backspace character is detected, remove the preceding character from
			// the input buffer, rewind the position for the next character, re-display
			if b[0] == 127 {
				if i > 0 {
					inputBuffer[i-1] = 0
					updateDisplay(true)
				}
				i -= 2
				// Correct buffer character position if required
				if i < 0 {
					i = -1
				}
			} else {
				// Append the typed character to the input buffer, and print it to the screen
				inputBuffer[i] = b[0]
				fmt.Printf("%c", b[0])
			}
			// End input collection of receipt of newline character
			if b[0] == '\n' {
				break
			}

		}
		// Move the contents of the input buffer into a slice
		var msgBuf []byte
		for i := 0; i < busSize; i++ {
			msgBuf = append(msgBuf, inputBuffer[i])
			if inputBuffer[i] == '\n' {
				break
			}
		}
		// Update display
		updateDisplay(false)
		// If the message buffer is empty (i.e. just a newline),
		// don't bother queuing a message to be sent - simply collect user input again
		if len(msgBuf) > 1 {
			// Handle name change command
			if string(msgBuf) == "name\n" {
				enableBufferingAndEcho()
				name = setName()
				disableBufferingAndEcho()
			} else {
				// Handle quit command
				if string(msgBuf) == "quit\n" {
					msgType = []byte("quit")
				}
				// Queue the message to be sent to the server
				msg := append(msgType, msgBuf...)
				msgBus.Enqueue(msg)
			}
		} else {
			fmt.Printf("\nMessage cannot be empty.\n%s: ", name)
		}
	}
}

// END DEFINITION

// DEFINE server setup functions

// Parses the command-line arguments into an IP address
// and Port number, and uses these as the connection address.
// If the IP or port are invalid, or if the connection is refused
// by the server, an error message is displayed and the client
// program terminates with status code 1.
func connectToServer() net.Conn {
	// Check first argument is valid IP address
	if net.ParseIP(os.Args[1]) == nil {
		fmt.Println("[USAGE]: ./nc $IP $Port")
		fmt.Println("Invalid IP address")
		os.Exit(1)
	}
	// Set host to be value of first argument
	host := os.Args[1]
	port := os.Args[2]
	// Banned due to strconv package
	// Set port number, exit if invalid
	// port, err := setPort(os.Args[2])
	// if err != nil {
	// 	fmt.Println("[USAGE]: ./nc $IP $Port")
	// 	fmt.Printf("Invalid port number. Must be between range %d - %d\n", portMin, portMax)
	// 	os.Exit(1)
	// }
	// Combine host and port to create server address string
	address := fmt.Sprintf("%s:%s", host, port)
	// Establish connection
	connection, err := net.DialTimeout(protocol, address, time.Second*5)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	connected = true
	return connection
}

// Banned due to 'strconv' package
// Parses a string to a port number
// func setPort(s string) (uint16, error) {
// 	port, err := strconv.ParseUint(s, 10, 16)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if port < uint64(portMin) {
// 		return 0, errors.New("invalid port number")
// 	}
// 	return uint16(port), nil
// }

// END DEFINITION

// DEFINE display and input functions

// Prints out the ASCII art TUX logo greeting message
func printGreeting() {
	fmt.Println("Welcome to TCP-Chat!")
	fmt.Println("         _nnnn_")
	fmt.Println("        dGGGGMMb")
	fmt.Println("       @p~qp~~qMb")
	fmt.Println("       M|@||@) M|")
	fmt.Println("       @,----.JM|")
	fmt.Println("      JS^\\__/  qKL")
	fmt.Println("     dZP        qKRb")
	fmt.Println("    dZP          qKKb")
	fmt.Println("   fZP            SMMb")
	fmt.Println("   HZM            MMMM")
	fmt.Println("   FqM            MMMM")
	fmt.Println(" __| \".        |\\dS\"qML")
	fmt.Println(" |    `.       | `' \\Zq")
	fmt.Println("_)      \\.___.,|     .'")
	fmt.Println("\\____   )MMMMMP|   .'")
	fmt.Println("     `-'       `--'")
}

// Prompts the user to enter their name, queues
// up a name request to be sent to the server, and returns
// the string corresponding to the name entered.
func setName() string {
	var result string
	var err error
	for {
		fmt.Printf("\n[ENTER YOUR NAME]: ")
		br := bufio.NewReader(os.Stdin)
		result, err = br.ReadString('\n')
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		if len(result) > 1 {
			break
		} else {
			fmt.Printf("\nName cannot be empty.\n")
		}
	}

	msgType := []byte("name")
	msg := append(msgType, []byte(result)...)
	msgBus.Enqueue(msg)
	result, _ = strings.CutSuffix(result, "\n")
	return result

}

// Updates the display with the latest chat history.
// Takes a boolean parameter which determines whether
// or not to print the input buffer.
func updateDisplay(withInput bool) {
	screenMutex.Lock()
	clear := exec.Command("clear")
	clear.Stdout = os.Stdout
	clear.Run()
	fmt.Printf("%s\n", history)
	fmt.Println("Enter 'name' to change your name")
	fmt.Printf("Enter 'quit' to disconnect - Using CTRL+C may affect terminal settings!\n")
	fmt.Printf("%s: ", name)
	if withInput {
		fmt.Printf("%s", inputBuffer)
	}
	screenMutex.Unlock()
}

func disableBufferingAndEcho() {
	// Disable input buffering and echo
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
}

func enableBufferingAndEcho() {
	// Restore input buffering and echo
	exec.Command("stty", "-F", "/dev/tty", "-cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "echo").Run()
}

// END DEFINITION
