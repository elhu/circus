package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const port = "6667"

func handleConn(c net.Conn) {
	log.Printf("%s connected", c.RemoteAddr().String())
	for {
		data, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			log.Println(fmt.Errorf("Error reading from socket: %v", err))
			break
		}
		command := strings.TrimSpace(data)
		if command == "STOP" {
			break
		} else {
			c.Write([]byte(command + "\n"))
		}
	}
	c.Close()
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Printf("Listening on port %s", port)

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(c)
	}
}
