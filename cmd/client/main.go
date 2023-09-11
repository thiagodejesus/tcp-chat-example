package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/thiagodejesus/tcp-chat-example/internal/conn"

	"github.com/google/uuid"
)

const address = "localhost:8080"

func main() {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	fmt.Println("\\q or quit to exit")
	fmt.Println("\\all [message] to send a message for all active connections")
	fmt.Print("\\list to list all active connections\n\n")

	fmt.Println("Sending dial to address:" + address)
	connection, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	establishedConnection := confirmConnection(connection)
	messageCh := make(chan string)
	go conn.HandleReceivedMessages(messageCh, establishedConnection.Connection, false, conn.ClientConn)
	go func() {
		for message := range messageCh {
			println(message)
		}
	}()

	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan() // use `for scanner.Scan()` to keep reading
		line := scanner.Text()
		if line == "quit" || line == "\\q" {
			fmt.Println("Finishing")
			break
		}
		moveCursorUp(1)
		clearLine(len(line))
		establishedConnection.WriteMessage(line, conn.ClientConn)
	}

	fmt.Println("Closing program by context")
	cancel()
	fmt.Println("Closing connection")
	connection.Close()
}

func clearLine(lineLength int) {
	// Create a string with spaces to overwrite the line
	clearString := strings.Repeat(" ", lineLength)

	// Move the cursor to the beginning of the line and overwrite it
	fmt.Printf("\r%s\r", clearString)
}

func moveCursorUp(n int) {
	fmt.Printf("\x1b[%dA", n) // ANSI escape code to move cursor up
}

// Must handle server closing

// Get Id, associate nick name and return a LiveConnection
func confirmConnection(connection net.Conn) conn.LiveConnection {
	messageCh := make(chan string)
	go conn.HandleReceivedMessages(messageCh, connection, true, conn.ClientConn)
	fmt.Println("Getting Id")
	var id uuid.UUID
	for message := range messageCh {
		id = uuid.MustParse(message[3:])
	}

	fmt.Printf("Connection established with id: %v\n", id)
	return conn.LiveConnection{
		Id:         id,
		Connection: connection,
	}
}
