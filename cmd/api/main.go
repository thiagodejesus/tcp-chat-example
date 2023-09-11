package main

import (
	"fmt"
	"log"
	"net"

	"github.com/thiagodejesus/tcp-chat-example/cmd/api/connmg"
	"github.com/thiagodejesus/tcp-chat-example/internal/conn"

	"github.com/google/uuid"
)

const (
	PORT = "8080"
)

func main() {
	address := "127.0.0.1:" + PORT
	fmt.Println("Address", address)

	l, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	activeConnections := connmg.ActiveConnections{
		Connections: make(connmg.ActiveConnectionsMap),
	}

	for {
		// Wait for a connection.
		connection, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		liveConnection := conn.LiveConnection{
			Id:         uuid.New(),
			Connection: connection,
		}

		fmt.Printf("[%v] New connection being handled\n", liveConnection.Id.String())
		go connmg.HandleNewConnection(&liveConnection, &activeConnections)
	}
}

// Errors that happened, EOF, Broken pipe, connection reset by peer
