package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const port = "6667"

type Circus struct {
	WG           *sync.WaitGroup
	ShutdownChan chan struct{}
	Listener     net.Listener
	Clients      []net.Conn
}

func NewCircus() *Circus {
	return &Circus{
		WG:           &sync.WaitGroup{},
		ShutdownChan: make(chan struct{}),
		Clients:      make([]net.Conn, 0),
	}
}

func (c *Circus) shuttingDown() bool {
	select {
	case <-c.ShutdownChan:
		return true
	default:
		return false
	}
}

func (c *Circus) handleConn(conn net.Conn) {
	c.WG.Add(1)
	defer c.WG.Done()

	log.Printf("%s connected", conn.RemoteAddr().String())
	for {
		if c.shuttingDown() {
			break
		}

		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("Error reading from socket: %v\n", err)
			break
		}
		command := strings.TrimSpace(data)
		if command == "STOP" {
			break
		} else {
			conn.Write([]byte(command + "\n"))
		}
	}
	if err := conn.Close(); err != nil {
		fmt.Printf("Error closing connection :%v\n", err)
	}
}

func (c *Circus) acceptLoop() {
	c.WG.Add(1)
	defer c.WG.Done()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	c.Listener = listener
	log.Printf("Listening on port %s", port)

	for {
		if c.shuttingDown() {
			break
		}
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept new connection: %v\n", err)
			continue
		}
		c.Clients = append(c.Clients, conn)
		go c.handleConn(conn)
	}
}

func (c *Circus) shutdown(sigs chan os.Signal) {
	<-sigs

	log.Println("Initiating shutdown sequence")
	signal.Stop(sigs)
	close(sigs)
	close(c.ShutdownChan)

	for _, conn := range c.Clients {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error closing connection :%v\n", err)
		}
	}

	if c.Listener != nil {
		c.Listener.Close()
	}
	log.Println("Shutdown sequence over")
	c.WG.Done()
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	c := NewCircus()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	c.WG.Add(1)
	go c.shutdown(sigs)

	go c.acceptLoop()
	c.WG.Wait()
}
