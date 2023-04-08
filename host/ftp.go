package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	SERVER_HOST = "localhost"
	SERVER_TYPE = "tcp"
)

var SERVER_PORT = generatePortNumber()

/* =============== Host FTP Server/Client =================== */

func ftpServer() {
	/*
	   Starts up server using the host, port, and
	   protocol defined above. Once a client is connected,
	   the processClient() function is ran as a goroutine (new thread)
	*/
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer func(server net.Listener) {
		err := server.Close()
		if err != nil {
			fmt.Println("Cannot close server:", err.Error())
			os.Exit(1)
		}
	}(server)

	fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Client connected")
		go processFTPClient(connection)
	}
}

func processFTPClient(connection net.Conn) {
	/*
		Listens for commands sent from the client. If a command requires a data transfer, the server connects to
		the data line hosted by the client before calling the handleDataTransfer() function that takes appropriate actions
		based on the instruction received. Once the transfer is complete, it closes its end of the data line and waits
		for a new instruction from the client. This process continues until "QUIT" is received from the client.
	*/

	var (
		buffer   = make([]byte, 1024)
		dataHost string
		dataPort string
	)

	// Receives and stores host and port number for data connection
	messageLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("[Control] Error reading:", err.Error())
		return
	}
	bufferToString := string(buffer[:messageLen])
	dataHostAndPort := strings.Split(bufferToString, ":")
	dataHost = dataHostAndPort[0]
	dataPort = dataHostAndPort[1]

	for {
		// Reads and deconstructs client message
		messageLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("[Control] Error reading:", err.Error())
			return
		}
		command := string(buffer[:messageLen])

		if command != "QUIT" {
			if isValidCommand(command) {
				dataConnection, err := net.Dial(SERVER_TYPE, dataHost+":"+dataPort)
				if err != nil {
					return
				}
				fmt.Println(fmt.Sprintf("[Data] Connected to %s:%s", dataHost, dataPort))

				err = transferData(command, dataConnection)
				if err != nil {
					fmt.Println("[Data] Error executing instruction:", err.Error())
					return
				}

				err = dataConnection.Close()
				if err != nil {
					fmt.Println("[Data] Error closing connection to server:", err.Error())
					return
				}
				fmt.Println("[Data] Connection closed")
			}
		} else if command == "QUIT" {
			err = connection.Close()
			if err != nil {
				fmt.Println("[Control] Error closing connection to client:", err.Error())
				return
			}
			fmt.Println("Connection Ended by Client")
			break
		}
	}
}

func ftpClient() {
	/*
		Prompts user to connect to the server using the command 'CONNECT <server name/IP address>:<port>'.
		Upon successful connection, the user is able to send commands to the server. If a command requires a
		data transfer, a server is started on the client to act as the data connection. Once the FTP server has
		been connected to the data line, the handleDataTransfer() function is called and runs the appropriate logic
		based on the command sent. When the transfer is complete, the data connection is closed and the user is
		prompted to send another command. This loop continues until the client sends the "QUIT" command.
	*/
	var (
		command        string
		connectPattern = `^CONNECT ([a-zA-Z0-9\-\.]+:[0-9]+)$`
		DATA_PORT      = generatePortNumber()
	)

	for {
		fmt.Println("[FTP] Connect to a server:")

		scanner := bufio.NewScanner(os.Stdin)

		if scanner.Scan() {
			command = scanner.Text()
		}

		if matched, err := regexp.MatchString(connectPattern, command); err == nil && matched {
			splitCommand := strings.Split(command, " ")
			hostAndPort := strings.Split(splitCommand[1], ":")
			host := hostAndPort[0]
			port := hostAndPort[1]
			control := connectToServer(host, port)

			if control == nil {
				continue
			}

			// Send Host/Port info for data connection
			_, err := control.Write([]byte(SERVER_HOST + ":" + DATA_PORT))
			if err != nil {
				fmt.Println("Unable to write to server:", err.Error())
				return
			}
			// Interact with the server via commands
			for {
				scanner := bufio.NewScanner(os.Stdin)
				fmt.Println("[FTP] Enter a command:")

				if scanner.Scan() {
					command = scanner.Text()
				}

				if command != "QUIT" {
					if isValidCommand(command) {
						fmt.Println("[Data] Port Running on " + SERVER_HOST + ":" + DATA_PORT)
						server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+DATA_PORT)
						if err != nil {
							fmt.Println("[Data] Error listening:", err.Error())
							return
						}

						_, err = control.Write([]byte(command))
						if err != nil {
							fmt.Println("[Control] Error writing:", err.Error())
							return
						}

						dataConnection, err := server.Accept()
						if err != nil {
							fmt.Println("[Data] Error accepting client:", err.Error())
							return
						}

						err = retrieveData(command, dataConnection)
						if err != nil {
							fmt.Println("[Data] Error in data transfer:", err.Error())
							return
						}

						fmt.Println("[Data] Port Closing")
						err = dataConnection.Close()
						if err != nil {
							fmt.Println("[Data] Error closing dataConnection to client:", err.Error())
							return
						}

						err = server.Close()
						if err != nil {
							fmt.Println("[Data] Error closing server:", err.Error())
							return
						}
					}

				} else if command == "QUIT" {
					_, err := control.Write([]byte(command))
					if err != nil {
						fmt.Println("[Control] Error writing:", err.Error())
						return
					}
					break
				} else {
					fmt.Println("Invalid command. Try again")
				}
			}
			err = control.Close()
			if err != nil {
				fmt.Println("[Control] Error closing connection to server:", err.Error())
				return
			}
		} else if command == "exit" {
			os.Exit(0)
		} else {
			fmt.Println("Invalid Command. Use `CONNECT <server name/IP address> <server port>` to connect to " +
				"a server")
		}
	}
}

func isValidCommand(command string) bool {
	/*
		Returns true if a given command is valid.
	*/
	dataPattern := `^(RETR|STOR) ([a-zA-Z0-9\-_]+)(\.[a-z]+)?$`
	matched, err := regexp.MatchString(dataPattern, command)

	if command == "LIST" || matched && err == nil {
		return true
	}
	fmt.Println("Invalid command or incorrect format. (Make sure to include the filename for STOR and RETR)")
	return false
}

func retrieveData(instruction string, dataConnection net.Conn) error {
	buffer := make([]byte, 1024)

	if instruction == "LIST" {
		//	Read from data, print contents to terminal
		dataLength, err := dataConnection.Read(buffer)
		if err != nil {
			fmt.Println("[Data] Error reading from client:", err.Error())
			return err
		}
		dataToString := string(buffer[:dataLength])
		fmt.Println(dataToString)
	} else {
		splitInstruction := strings.Split(instruction, " ")
		command := splitInstruction[0]
		filename := splitInstruction[1]

		if command == "STOR" {
			//	Send file to the server
			file, err := os.Open("./" + filename)
			if err != nil {
				fmt.Println(err)
				return err
			}
			_, err = io.Copy(dataConnection, file)
			if err != nil {
				fmt.Println(err)
				return err
			}
			err = file.Close()
			if err != nil {
				fmt.Println(err)
				return err
			}
		} else if command == "RETR" {
			//	Retrieve file from the server
			file, err := os.Create(filename)
			if err != nil {
				fmt.Println("Error creating file:", err.Error())
				return err
			}
			_, err = io.Copy(file, dataConnection)
			if err != nil {
				fmt.Println("Error copying data: ", err.Error())
				return err
			}
			err = file.Close()
			if err != nil {
				fmt.Println("Error closing file: ", err.Error())
				return err
			}
		}
	}
	return nil
}

func transferData(instruction string, dataConnection net.Conn) error {
	/*
		Executes appropriate actions based on the command passed. Returns any potential errors.
	*/
	if instruction == "LIST" {
		// Build a string that contains all files in the current directory, send to client
		data := ""
		files, err := os.ReadDir(".")

		if err != nil {
			return err
		}

		for _, file := range files {
			data += file.Name() + " "
		}

		_, err = dataConnection.Write([]byte(data))
		if err != nil {
			fmt.Println("[Data] Error writing:", err.Error())
			return err
		}
	} else {
		splitInstruction := strings.Split(instruction, " ")
		command := splitInstruction[0]
		filename := splitInstruction[1]

		if command == "STOR" {
			//	Receive a file from the client
			file, err := os.Create(filename)
			if err != nil {
				fmt.Println("Error creating file:", err.Error())
				return err
			}
			_, err = io.Copy(file, dataConnection)
			if err != nil {
				fmt.Println("Error copying data: ", err.Error())
				return err
			}
			err = file.Close()
			if err != nil {
				fmt.Println("Error closing file: ", err.Error())
				return err
			}
		} else if command == "RETR" {
			// Send a file to the client
			file, err := os.Open("./" + filename)
			if err != nil {
				fmt.Println(err)
				return err
			}
			_, err = io.Copy(dataConnection, file)
			if err != nil {
				fmt.Println(err)
				return err
			}
			err = file.Close()
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

func generatePortNumber() string {
	rand.Seed(time.Now().UnixNano())

	port := rand.Intn(5000) + 5000

	return strconv.Itoa(port)
}
