# Peer-to-Peer Server

## Hosts/Peers
  * Accepts commands from users
  * Registers with centralized server by sending a list of file names and keywords
  * Can query the server with a keyword search
  * Downloads files from other hosts (NOT the centralized server)
  * Serves requested files to other hosts


## Central Server
  * Tracks current users and the files they share
  * Once a connection is established, store username, connection speed (made up), and hostname in a "users" table. 
  * Store file descriptions in a "files" table
  * Provide keyword search to users to query the "files" table that returns the resource location(s) of the remote file(s) matching the search

## To-Do:
  *  