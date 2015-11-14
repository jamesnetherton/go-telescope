package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/tarm/serial"
)

func main() {
	bindAddress := flag.String("bind-address", "0.0.0.0", "A host name or ip address for the server to bind to")
	port := flag.String("port", "9595", "A port for the server to listen on")
	scopeDevice := flag.String("device", "", "The device that the telescope should connected to")
	flag.Parse()

	socketChan := make(chan []byte)
	scopeChan := make(chan []byte)

	c := &serial.Config{Name: *scopeDevice, Baud: 9600}
	s, err := serial.OpenPort(c)

	if err == nil {
		defer s.Close()
		go readFromTelescope(s, socketChan)
		go writeToTelescope(s, scopeChan)
	} else {
		log.Fatal("Failed to connect to the telescope")
	}

	ln, err := net.Listen("tcp", *bindAddress+":"+*port)
	if err == nil {
		defer ln.Close()

		for {
			conn, err := ln.Accept()

			if err == nil {
				go readFromSocket(conn, scopeChan)
				go writeToSocket(conn, socketChan)
			}
		}
	} else {
		log.Fatalf("Failed to connect to %s:%s", *bindAddress, *port)
	}
}

func readFromSocket(c net.Conn, channel chan []byte) {
	for {
		buf := make([]byte, 1024)

		bytesRead, err := c.Read(buf)
		if err == nil {
			channel <- buf[:bytesRead]
		} else {
			return
		}
	}
}

func writeToSocket(c net.Conn, channel chan []byte) {
	for {
		data := <-channel

		_, err := c.Write(data)

		if err != nil {
			fmt.Println(err)
			c.Close()
			return
		}
	}
}

func readFromTelescope(s *serial.Port, channel chan []byte) {
	for {
		buf := make([]byte, 1024)
		bytesRead, _ := s.Read(buf)
		fmt.Println(buf[:bytesRead])
		channel <- buf[:bytesRead]
	}
}

func writeToTelescope(s *serial.Port, channel chan []byte) {
	for {
		data := <-channel
		fmt.Println(data)
		s.Write(data)
		s.Flush()
	}
}
