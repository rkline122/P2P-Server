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
	files  = make([]FileEntry, 0)
	buffer = make([]byte, 1024)
)

type FileEntry struct {
	owner           string
	ftpServerAddr   string
	connectionSpeed string
	fileName        string
	description     string
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
		go processClient(connection)
	}
}

func processClient(connection net.Conn) {
	var (
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
	ftpAddr := fmt.Sprintf("%s:%s", clientHostname, clientPort)

	fmt.Printf("%s has connected\n", clientUsername)

	// Read and store file descriptions in global slice
	mLen, err = connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}
	bufferToString = string(buffer[:mLen])
	lines := strings.Split(bufferToString, "\n")

	for i := 0; i < len(lines)-1; i++ {
		splitLine := strings.Split(lines[i], ", ")
		fileName := splitLine[0]
		description := splitLine[1]

		entry := FileEntry{
			owner:           clientUsername,
			ftpServerAddr:   ftpAddr,
			connectionSpeed: clientSpeed,
			fileName:        fileName,
			description:     description,
		}
		files = append(files, entry)
	}

	// Handle Keyword searches
	handleKeywordSearch(connection)
	fmt.Printf("%s disconnected\n", clientUsername)
	connection.Close()

}

func handleKeywordSearch(connection net.Conn) {
	for {
		var searchResults string

		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}
		bufferToString := string(buffer[:mLen])
		fmt.Printf("%s\n", bufferToString)

		if bufferToString == "quit" {
			break
		} else {
			results := filterByKeyword(bufferToString)
			if len(results) > 0 {
				for _, file := range results {
					fileStr := fmt.Sprintf("Filename: %s | Description: %s | Host: %s | Connection Speed: %s\n", file.fileName, file.description, file.ftpServerAddr, file.connectionSpeed)
					searchResults += fileStr
				}
			} else {
				searchResults = "No files found matching search"
			}
			fmt.Println(searchResults)
			connection.Write([]byte(searchResults))
		}
	}
}

func filterByKeyword(keyword string) []FileEntry {
	var matches []FileEntry

	for _, entry := range files {
		if strings.Contains(entry.description, keyword) {
			matches = append(matches, entry)
		}
	}
	return matches
}
