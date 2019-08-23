package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"time"
)

// SyncFolder defines where the savegames will reside
const SyncFolder string = "/tmp/minesync"

type syncObject struct {
	Name string
	Data []byte
}

type saveList struct {
	Saves []save
}

type save struct {
	Name             string
	LastModifiedDate time.Time
}

func main() {
	go startSyncServer()
	go startDownloadServer()
	startResourceServer()
}

// Set the server and listen for incoming requests
func startResourceServer() {
	ln, err := net.Listen("tcp", ":9998")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening for connections on 9998...")
	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleResourceServerConnection(c)
	}
}

// Handles the listing of savegames
func handleResourceServerConnection(c net.Conn) {
	files, err := ioutil.ReadDir(SyncFolder)
	if err != nil {
		fmt.Println(err)
		return
	}

	saves := make([]save, 0)
	for _, f := range files {
		if filepath.Ext(SyncFolder+"/"+f.Name()) == ".zip" {
			saves = append(saves, save{f.Name(), f.ModTime()})
		}
	}

	err = gob.NewEncoder(c).Encode(saveList{saves})
	if err != nil {
		fmt.Println(err)
	}
}

// Set the server and listen for incomming connections
func startSyncServer() {
	ln, err := net.Listen("tcp", ":9999")
	defer ln.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening for connections on 9999...")
	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleSyncServerConnection(c)
	}
}

// Handle the client connection
func handleSyncServerConnection(c net.Conn) {
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

func startDownloadServer() {
	ln, err := net.Listen("tcp", ":9997")
	defer ln.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening for connections on 9997...")
	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleDownloadServerConnection(c)
	}
}

func handleDownloadServerConnection(c net.Conn) {
	defer c.Close()

	decoded := save{}
	gob.NewDecoder(c).Decode(&decoded)

	s := syncObject{}
	f, err := ioutil.ReadFile(SyncFolder + "/" + decoded.Name)
	if err != nil {
		fmt.Println(err)
		return
	}

	str := ""
	for _, ss := range strings.Split(decoded.Name, "_") {
		if ss != "minesync" {
			str += ss + " "
		}
	}
	s.Name = strings.TrimSpace(strings.Replace(str, ".zip", "", -1))
	s.Data = f

	gob.NewEncoder(c).Encode(s)
}
