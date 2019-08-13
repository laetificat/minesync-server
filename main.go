package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
)

// SyncFolder defines where the savegames will reside
const SyncFolder string = "/tmp/minesync"

type syncObject struct {
	Name string
	Data []byte
}

func main() {
	server()
}

// Set the server and listen for incomming connections
func server() {
	ln, err := net.Listen("tcp", ":9999")
	defer ln.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening for connections...")
	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleServerConnection(c)
	}
}

// Handle the client connection
func handleServerConnection(c net.Conn) {
	defer c.Close()

	decoded := syncObject{}
	gob.NewDecoder(c).Decode(&decoded)

	_, err := ioutil.ReadDir(SyncFolder)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile(SyncFolder+"/"+decoded.Name, decoded.Data, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Received", decoded.Name)
}
