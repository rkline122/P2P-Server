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
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "8636"
	SERVER_TYPE = "tcp"
)

var (
	files = make([]FileEntry, 0)
)

type FileEntry struct {
	name        string
	description string
}

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

	// exampleFile := FileEntry{
	// 	name:        "example",
	// 	description: "example description",
	// }
	// files = append(files, exampleFile)

	mLen, err = connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}
	bufferToString = string(buffer[:mLen])

	fmt.Println(bufferToString)

	// for _, val := range files {
	// 	fmt.Println(val)
	// }
}

func parseFileDescriptions(fileName string) {
	// Parses file and stores file names and descriptions in global slcie
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Split the line by comma to extract the filename and file description
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			fmt.Println("Failed to parse line:", line)
			continue
		}
		filename := strings.TrimSpace(parts[0])
		fileDescription := strings.TrimSpace(parts[1])

		// Create a FileEntry struct and append it to the slice
		fileEntry := FileEntry{
			name:        filename,
			description: fileDescription,
		}
		files = append(files, fileEntry)

	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Failed to read file:", err)
		return
	}

}
