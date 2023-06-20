# Net-Cat

*N.B. When running the client application, please issue the 'quit' command to
disconnect from the server, rather than CTRL+C. For more details as to why,
please read the 'Disabling input buffering and terminal echo' section at the
end of this README.*

Netcat is a simple terminal-based group chat application. Composed of a
server program and one or more instances of the client program written in Go,
it utilises TCP sockets (version 4) and IP as the method of communication.

## How to build and run

You can build and run either the client, the server, or both, depending on
what scenario you want to achieve.

To build the server program, change directory into src/server and issue the
command 'go build -o TCPChat main.go'.

To run the server, from within the src/server directory, issue the command
'./TCPChat $Port' ($Port is an optional command-line argument corresponding
to a port number for the server to listen on. Omitting it defaults to port
8989.

To build the client program, change directory into src/client and issue the
command 'go build -o nc main.go'.

To run the client, from within the src/client directory, issue the command
'./nc $IP $Port', where $IP is replaced with the IP address of the server
e.g. 127.0.0.1 for localhost), and $Port is replaced with the server's port
number.

## Client User Commands

The user of the client program is prompted to send messages to the chat.
The two following commands have special behaviour. Any other user input is
treated as the contents of a message to be sent to the group chat (or, in 
the case of a name prompt, the name for the client user).

- "name": If the user types in "name" when prompted to send a message, they
are redirected to update the name associated with their client instance.

- "quit": If the user types in "quit" when prompted to send a message, they
inform the server that they are disconnecting, and the client closes its
socket connection to the server.

## Message Protocol

*N.B. The user of the client program does not need to manually enter the
message type. By default, the program will add the correct prefix to all
outgoing messages. The user can issue the commands 'quit' and 'name', to
disconnect from the server and change their name respectively. All other user
input is treated as the contents of a message to be sent to the group chat.*
 
Messages are sent over the socket connection as a stream ('slice') of bytes.
In order to differentiate between different types of client requests, we
define the content of the first 4 bytes (converted to a string) to represent
the 'type' of the client request, as follows:

- "name": Any message received from the client with the prefix "name"
represents a request to set the name associated with that client. The server
responds with simply a null-byte (i.e. []byte{0}).

- "hist": Represents a request for the latest chat history. The server
responds with all the bytes read from the file, null-byte terminated.

- "text": Represents a request to send a message to the group chat with the
content that follows. The server responds with a null-byte.

- "quit": Represents a request to disconnect from the server. The server does
not respond, but instead gracefully closes the connection server-side, since
the client also disconnects immediately upon sending a quit request.

## Handling maximum number of client connections

The server program ensures a maximum of 10 simultaneous client connections by
closing its listener when 10 clients are connected to the server. It re-opens
the listener and accepts connections again once a client disconnects.

## Message Logging and Chat History

The server logs all messages intended for the group chat (i.e. those with the
type "text") to a text file called "history.txt". Similarly, when the server
receives a request for the latest chat history from a client (type "hist"),
it reads the contents from this file to send as its response. Since separate
go-routines will access this file, each go-routine must first acquire a mutex
lock before commencing a read/write operation. It then releases the lock once
the operation is complete. This ensures that the file is concurrency-safe,
and will not carry out any unexpected behaviour such as mixing messages or
reading partially logged messages.

## Use of a Client 'Message Bus' for concurrency

Upon starting an instance of the client program, a Message Bus is intialised
with a Queue (a []byte channel), to which all outgoing client messages are
sent. This allows for a design in which different go-routines manage the
different responsibilities of collecting user input; periodically queuing a
request for the latest chat history; and communicating with the server.

## Disabling input buffering and terminal echo

Once a client has connected and the user has set its name, the client program
disables input buffering and terminal echo. This allows the go-routine whose
responsibility is to fetch the latest chat history to clear the screen and
print updates without losing the display of any input that may have been
typed but not yet sent to the server.

As a result of this, terminating the client program with an interrupt signal
(CTRL+C) is not advised, since the original terminal settings (buffered input
and terminal echo) are not restored. Therefore, a message advising the user
to quit the program by issuing a 'quit' command precedes the input prompt. 
