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
)

func connectToServer(host, port string) net.Conn {
	connection, err := net.Dial(SERVER_TYPE, host+":"+port)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(fmt.Sprintf("[Control] Connected to %s:%s", host, port))
		return connection
	}
	return nil
}

/* =============== Host to Central Server ================== */

func registerWithServer() {
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
			/* Send list of files and keywords */
			/* Query with keyword search */
			/* Connect with other hosts via FTP */

		} else if command == "exit" {
			os.Exit(0)
		} else {
			fmt.Println("Invalid Command. Use `CONNECT <server name/IP address> <server port>` to connect to a server")
		}

	}
}

func getHostInfo() string {
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

func main() {
	registerWithServer()
}