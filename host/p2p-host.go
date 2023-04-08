/*
Project 3: P2P Server
By Ryan Kline
	---
CIS 457 - Data Communications
Winter 2023
=====================
Host
*/

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

func connectToServer(host, port string) net.Conn {
	/*
		Given a host and port, a connection is established.
		Returns an interface of type net.Conn if successful.
	*/
	connection, err := net.Dial(SERVER_TYPE, host+":"+port)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(fmt.Sprintf("Connected to %s:%s", host, port))
		return connection
	}
	return nil
}

func registerWithServer() {
	/*
		Connects the host to the central server and sends the server
		the hosts' information (username, hostname, connection speed, etc.).

		Upon successful connection, the host is able to enter commands to query
		the sever, connect with other hosts, or to terminate this connection.
	*/
	var (
		connectPattern = `^CONNECT ([a-zA-Z0-9\-\.]+:[0-9]+)$`
		command        string
	)

	for {
		fmt.Println("Connect to the Central Server:")
		scanner := bufio.NewScanner(os.Stdin)

		if scanner.Scan() {
			command = scanner.Text()
		}

		if matched, err := regexp.MatchString(connectPattern, command); err == nil && matched {
			splitCommand := strings.Split(command, " ")
			hostAndPort := strings.Split(splitCommand[1], ":")
			host := hostAndPort[0]
			port := hostAndPort[1]

			connection := connectToServer(host, port)

			if connection == nil {
				continue
			}

			/* Send username, hostname, port, and connection speed */
			_, err = connection.Write([]byte(getHostInfo()))
			if err != nil {
				fmt.Println("Unable to write to server:", err.Error())
				return
			}

			sendFileDescriptor("filelist.txt", connection)

			fmt.Println("Connection successful!")
			for {
				printCommands()
				fmt.Println("Enter a command:")
				scanner := bufio.NewScanner(os.Stdin)

				if scanner.Scan() {
					command = scanner.Text()
				}

				if command == "search" {
					keywordSearch(connection)
				} else if command == "ftp" {
					ftpClient()
				} else if command == "quit" {
					fmt.Println("Terminating connection")
					break
				} else {
					fmt.Println("Invalid command. Try again")
				}

			}
			connection.Close()

		} else if command == "exit" {
			os.Exit(0)
		} else {
			fmt.Println("Invalid Command. Use `CONNECT <server name/IP address> <server port>` to connect to a server")
		}

	}
}

func printCommands() {
	/*
		Prints a list of available commands to the terminal
	*/
	fmt.Println("\nAvailable Commands:")
	fmt.Println("'search' - submit a query for files on the server by their descriptions")
	fmt.Println("'ftp' - initialize a ftp connection with another host on the server")
	fmt.Println("'quit' - terminate the connection to the server.\n")
}

func keywordSearch(connection net.Conn) {
	/*
		Gets input from the user and sends it to the server.
		Prints the results recieved by the server
	*/
	var input string
	buffer := make([]byte, 1024)

	for {
		fmt.Println("Enter a keyword to search for:")
		scanner := bufio.NewScanner(os.Stdin)

		if scanner.Scan() {
			input = scanner.Text()
		}
		break
	}

	connection.Write([]byte(input))
	time.Sleep(2)
	mLen, _ := connection.Read(buffer)
	bufferToString := string(buffer[:mLen])
	entries := strings.Split(bufferToString, "\n")

	fmt.Println("\nSearch Results:")
	for _, entry := range entries {
		fmt.Printf("%s\n", entry)
	}
}

func getHostInfo() string {
	/*
		Returns a string of the hosts information (username, hostname, port, connection speed)
	*/
	var (
		username        string
		connectionSpeed string
		hostname        = SERVER_HOST
		port            = SERVER_PORT
	)

	for {
		fmt.Println("Enter your username:")
		scanner := bufio.NewScanner(os.Stdin)

		if scanner.Scan() {
			username = scanner.Text()
		}

		if username == "" {
			fmt.Println("Cannot have an empty username")
			continue
		} else {
			break
		}
	}

	for {
		fmt.Println("Enter your connection speed (slow, medium, fast):")
		scanner := bufio.NewScanner(os.Stdin)

		if scanner.Scan() {
			connectionSpeed = scanner.Text()
		}

		strings.TrimSpace(connectionSpeed)
		if !(connectionSpeed == "slow" || connectionSpeed == "medium" || connectionSpeed == "fast") {
			fmt.Println("Your input:", connectionSpeed)
			fmt.Println("Invalid connection speed")
			continue
		} else {
			break
		}
	}

	info := fmt.Sprintf("%s %s %s %s", username, hostname, port, connectionSpeed)

	return info

}

func sendFileDescriptor(fileName string, connection net.Conn) {
	/*
		Parses the file descriptor and writes its contents to the server
	*/
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	var fileStr string = ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fileStr += line + "\n"
	}

	_, err = connection.Write([]byte(fileStr))
	if err != nil {
		fmt.Println("Unable to write to server:", err.Error())
		return
	}
}

func main() {
	go ftpServer()
	registerWithServer()
}