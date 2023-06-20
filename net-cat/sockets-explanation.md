# Sockets Programming

'Sockets' are a software abstraction of a direct connection between two or more computer processes. We are going to make two basic Go programs, a 'server' and a 'client', which use a 'protocol' to communicate over a 'network'. A user of the client program will type in messages to send to the server, and the server will simply echo back to the client the message it received. 

## Protocol - What is it?

Socket connections are established with a communication protocol (such as TCP, UDP, etc.) and the network address of the process to which it is connecting.

The protocol is simply a well-defined method of communication. In every day life, when we communicate with each other through speech or writing, we implicitly decide on a 'protocol' by speaking the same language. Say I opened a shop to sell cars (I am the 'server' - I am offering a service). I don't speak German, so if a 'client' started trying to make an offer to me in German, this is analogous to a client computer attempting to communicate with a server using a protocol which is not supported. Protocols may contain restrictions such as on the format of messages; message length; and the contents of a message (analogy: Someone tries to buy fruit from my car shop... like mate, what are you doing?). In this way, the communication between client and server is well-defined, and facilitates building programs which make use of it.

### TCP (Transmission Control Protocol) and IP (Internet Protocol)

TCP is a common protocol for direct communication between socket connections. IP is the protocol for communication between processes over the internet (no more details here, but you can look it up if you're interested).

When using TCP/IP, the network address of a process is uniquely identified by an IP address and a port number. For the sake of this explanation, we will concern ourselves with TCP and IP version 4. (Version 6 is available, but version 4 is still very widely used, and the idea is the same).

For example, to set up a client program to communicate with a server process running at the IP address '192.198.62.88', on the port '5555', the complete network address is given by '192.168.62.88:5555'. If we get the IP address or the port number wrong, we will in all likelihood be unable to communicate with the server (unless, of course, we happen to stumble upon another process which is listening for TCP connections at the IP address and port we chose).

As an analogy, treat a computer as a block of flats. Its IP address is the address of the building, and its port numbers are the numbers of the individual flats. If we are sending a package to our friend at their flat - it's no good if we get the wrong building... it's also no good if we get the right building but the wrong flat number, or indeed the wrong building with the same flat number. In all these cases, our package will go to the wrong location.

## Client-Server Communication

Ok, so we get that in order to create socket connections, we need to choose a protocol (let's choose TCP), and we know that an IP address and port number together define the address of a computer process. What now?

Our end goal is to have the client and server send messages back and forth to each other. To simplify what we are trying to do, it is useful to think of both the client and the server as one person on each side of a seesaw. When one person goes up, the other must go down, and vice versa. When the client sends a message, the server must receive the message. When the server sends a response, the client must receive the response, and so on.

So, considering this, let's think about what we need to build this communication 'seesaw'. Well, we need to have a server responsible for 'serving' some information. To activate a server, we need to give it a socket, or open a connection, at a specific port, using a specific protocol, and then we need to allow that socket to 'listen' out for any clients who wish to connect.

Great. Let's say we create our socket connection with TCPv4 on port 5555. Just as our server's first task was to listen out for connections, our client's first task should be to attempt to connect to said server. We create the socket on the client to also use TCPv4, and then we supply it with the network address of the server (its IP address, followed by 5555). The client is going to 'dial' the server in an attempt to connect.

Next, the server must 'accept' any incoming connection requests.

Now, once a connection has been established between client and server, we must make a decision on who gets to 'write' the first message. The other gets to 'read' the first message (read/write are like the two sides of the seesaw). Let's say we want the client to send the first message - maybe the name of the person using the client program. Therefore, the client is going to 'write' or 'send', while the server is going to 'read' or 'receive' some information.

Once the client has sent its message, now it is its turn to 'read' or 'receive' the response from the server. Meanwhile, once the server has received the message from the client, it is now its turn to 'write' or 'send' its response back to the client.

For the case of this exercise, let's make the server simply echo back to the client whatever message it sent, and have it display on the client. We will use os.Stdin (i.e. the terminal) to get user input for the contents of the client message. One final thing to note: we will store the messages in an array or 'slice' of bytes - since, after all, this is exactly how computers store information.

## Summary of Steps

- Server: Create a socket connection, specifying a protocol and a port number.
- Client: Create a socket connection, specifying a protocol, and the network address (IP address:port) of the server.
- Server: Listen out for incoming client connections
- Client: Dial the server (attempt to connect)
- Server: Accept incoming client connection
- Loop: Client writes message/Server reads message... Server writes message/Client reads message