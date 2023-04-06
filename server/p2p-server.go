/*
Project 3: P2P Server
By Ryan Kline
	---
CIS 457 - Data Communications
Winter 2023
=====================
Centralized Server
*/

package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "8636"
	SERVER_TYPE = "tcp"
)

func main() {
	fmt.Println("Server Running...")
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
	fmt.Println("Waiting for client...")

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Client connected")
		go processClient(connection)
	}
}

func processClient(connection net.Conn) {
	var (
		buffer         = make([]byte, 1024)
		clientUsername string
		clientHostname string
		clientPort     string
		clientSpeed    string
	)

	// Read host info
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}
	bufferToString := string(buffer[:mLen])
	hostInfo := strings.Split(bufferToString, " ")
	clientUsername = hostInfo[0]
	clientHostname = hostInfo[1]
	clientPort = hostInfo[2]
	clientSpeed = hostInfo[3]

	tmpStr := fmt.Sprintf("User: %s has connected with a speed of %s."+
		"\nTheir hostname is: %s and are listening on port %s for FTP connections.", clientUsername, clientSpeed, clientHostname, clientPort)
	fmt.Println(tmpStr)

	// Recieve host description file
	file, err := os.Create(fmt.Sprintf("%s_file_descriptions.txt", clientUsername))
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer file.Close()

	// Copy the data from the network connection to the file
	_, err = io.Copy(file, connection)
	if err != nil {
		fmt.Println("Failed to receive file:", err)
		return
	}
}