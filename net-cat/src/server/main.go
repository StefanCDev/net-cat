package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// DEFINE Program constants
const (
	// defaultPort          uint16 = 8989
	defaultPort          string = "8989"
	host                 string = "192.168.1.136"
	maxClientConnections int    = 10
	msgPrefixSize        int    = 4
	portMin              uint16 = 1000
	portMax              uint16 = 65535
	protocol             string = "tcp4"
)

// END DEFINITION

var namesMap map[net.Conn]string = make(map[net.Conn]string)
var fileMutex sync.Mutex
var connections int = 0

func main() {
	var port string
	// Check number of arguments. If no port is supplied,
	// set to defaultPort. If there are too many args, exit
	switch len(os.Args) {
	case 1:
		port = defaultPort
	case 2:
		port = os.Args[1]
		// Banned due to 'strconv' package
		// var err error
		// port, err = setPort(os.Args[1])
		// if err != nil {
		// 	fmt.Println("[USAGE]: ./TCPChat $port")
		// 	os.Exit(1)
		// }
	default:
		fmt.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	}
	// Listen from address created with host and port number supplied as argument
	address := fmt.Sprintf("%s:%s", host, port)
	listener, err := net.Listen(protocol, address)
	if err != nil {
		fmt.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	}
	fmt.Printf("Listening on port %s...\n", port)
	// Define a boolean representing whether the listener is closed
	closed := false
	for {
		// If the number of connections reaches limit,
		// close the listener, and wait 5 seconds before next loop
		// iteration
		if connections >= maxClientConnections {
			listener.Close()
			closed = true
			time.Sleep(5 * time.Second)
		} else {
			// If the listener is closed, but there are fewer connections
			// than the limit, listen again for TCP connections
			if closed {
				closed = false
				listener, err = net.Listen(protocol, address)
			}
			exitOnError(err)
			// Accept client connection
			connection, err := listener.Accept()
			exitOnError(err)
			// Increment number of connections
			connections++
			fmt.Printf("Client connected\nNumber of connected clients: %d\n", connections)
			// A dedicated go-routine handles all communication with this client
			go connectionHandler(connection)
		}

	}
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Banned due to usage of 'strconv' package
// This parses the port number supplied as a command line argument
// func setPort(s string) (uint16, error) {
// 	port, err := strconv.ParseUint(s, 10, 16)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if port < uint64(portMin) {
// 		errMsg := fmt.Sprintf("invalid port number '%d'. must be in range %d - %d", port, portMin, portMax)
// 		return 0, errors.New(errMsg)
// 	}
// 	return uint16(port), nil
// }

func connectionHandler(connection net.Conn) {
	defer connection.Close()
	// Create a buffered reader to read from client socket connection
	br := bufio.NewReader(connection)
	// Declare name variable
	var name string
	connected := true
	for connected {
		// Read up to and including the newline character from the socket stream
		buffer, err := br.ReadBytes('\n')
		if err != nil {
			// Any problem with the buffered reader,
			// disconnect the client. Set connected to false,
			// decrement number of connections, and delete the name
			// associated with the connection from the map.
			connected = false
			connections--
			fmt.Printf("%s disconnected badly\n", name)
			fmt.Printf("Number of connected clients: %d\n", connections)
			delete(namesMap, connection)
			// Log a welcome message to the chat history
			message := fmt.Sprintf("*** %s has left the chat ***\n", name)
			writeToChat(message)
			return
		}
		// Slice the byte stream to retrieve the message type
		// and the message contents
		msgType := string(buffer[0:msgPrefixSize])
		contents := string(buffer[msgPrefixSize:])
		switch msgType {
		// Request for latest chat history
		case "hist":
			// fmt.Printf("%s - Chat history request\n", name)
			history := readHistory()
			history = append(history, 0)
			_, err := connection.Write(history)
			exitOnError(err)
		// Request to disconnect from the server
		case "quit":
			// Log disconnection, set connected to false,
			// decrement number of connections, and delete the
			// name associated with the connection from the map
			fmt.Printf("%s disconnected gracefully\n", name)
			connected = false
			connections--
			delete(namesMap, connection)
			fmt.Printf("Number of connected clients: %d\n", connections)
			// Log a leaving message to the chat history
			message := fmt.Sprintf("*** %s has left the chat ***\n", name)
			writeToChat(message)
			// go routine returns here, since client has disconnected
			return
		// Request to create/change name of client
		case "name":
			fmt.Println("Name Request")
			name, _ = strings.CutSuffix(contents, "\n")
			// Get old name, if it exists
			oldName, found := namesMap[connection]
			namesMap[connection] = name
			var message string
			// If an old name exists, the request is to change the name
			// Otherwise, it is a name definition, and the client has just joined.
			if found {
				message = fmt.Sprintf("*** %s changed their name to %s ***\n", oldName, name)
			} else {
				message = fmt.Sprintf("*** %s joined the chat ***\n", name)
			}
			writeToChat(message)
			response := []byte{0}
			_, err := connection.Write(response)
			exitOnError(err)
		// Request to send a message to the group chat
		case "text":
			fmt.Printf("%s - Message Request\n", name)
			dateTimeStr := buildDateTimeString(time.Now())
			// Log client message with timestamp and name to chat history
			message := fmt.Sprintf("[%s][%s]: %s", dateTimeStr, name, contents)
			writeToChat(message)
			response := []byte{0}
			_, err := connection.Write(response)
			exitOnError(err)
		}
	}
}

// This function acquires a mutex lock on the history file,
// opens it (creates if it does not exist) and appends the
// message supplied as a string argument. It then releases the mutex lock.
func writeToChat(msg string) {
	fileMutex.Lock()
	fp, err := os.OpenFile("history.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		fileMutex.Unlock()
		log.Fatal(err.Error())
	}
	defer fp.Close()
	fmt.Fprint(fp, msg)
	fileMutex.Unlock()
}

// This function acquires a mutex lock on the chat history file,
// and reads in all bytes. It then appends a null-byte to the result,
// releases the mutex lock, and returns the result
func readHistory() []byte {
	fileMutex.Lock()
	result, err := os.ReadFile("history.txt")
	if err != nil {
		fileMutex.Unlock()
		log.Fatal(err.Error())
	}
	fileMutex.Unlock()
	return result
}

// This function builds a date-time string, to be
// prepended before the name of a message author and
// the contents of the message.
func buildDateTimeString(t time.Time) string {
	tYear := t.Local().Year()
	tMonth := t.Local().Month()
	tMonthStr := fmt.Sprintf("%d", int(tMonth))
	if tMonth < 10 {
		tMonthStr = "0" + tMonthStr
	}
	tDay := t.Local().Day()
	tDayStr := fmt.Sprintf("%d", tDay)
	if tDay < 10 {
		tDayStr = "0" + tDayStr
	}
	dateStr := fmt.Sprintf("%d-%s-%s", tYear, tMonthStr, tDayStr)
	tHour := t.Local().Hour()
	tHourStr := fmt.Sprintf("%d", tHour)
	if tHour < 10 {
		tHourStr = "0" + tHourStr
	}
	tMin := t.Local().Minute()
	tMinStr := fmt.Sprintf("%d", tMin)
	if tMin < 10 {
		tMinStr = "0" + tMinStr
	}
	tSec := t.Local().Second()
	tSecStr := fmt.Sprintf("%d", tSec)
	if tSec < 10 {
		tSecStr = "0" + tSecStr
	}
	timeStr := fmt.Sprintf("%s:%s:%s", tHourStr, tMinStr, tSecStr)
	return fmt.Sprintf("%s %s", dateStr, timeStr)
}
