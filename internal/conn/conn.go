package conn

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

var mu sync.Mutex

type ConnEnv int

const (
	ServerConn ConnEnv = iota
	ClientConn
)

func (s ConnEnv) String() string {
	switch s {
	case ServerConn:
		return "server"
	case ClientConn:
		return "client"
	}
	return "unknown"
}

type LiveConnection struct {
	Id         uuid.UUID
	Connection net.Conn
}

func (lc *LiveConnection) WriteMessage(message string, connEnv ConnEnv) {
	messageLength := len([]rune(message))

	messageParsed := fmt.Sprintf("%vmLen"+message, messageLength)
	_, err := lc.Connection.Write([]byte(messageParsed))
	if errors.Is(err, syscall.EPIPE) {
		fmt.Println("Broken pipe")
	}
	if err != nil {
		if connEnv.String() == ClientConn.String() {
			fmt.Println("Connection closed by server")
			log.Fatal(err)
		}

		if errors.Is(err, syscall.EPIPE) {
			fmt.Println("Broken pipe")
			return
		}

		log.Fatal(err)
	}
}

func HandleReceivedMessages(messageCh chan string, conn net.Conn, isSingleMessage bool, connEnv ConnEnv) {
	defer close(messageCh)
	messageSize := 0
	message := ""
	bufferSize := 1024
	buffer := make([]byte, bufferSize)

	for {
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			if connEnv.String() == ClientConn.String() {
				fmt.Println("Connection closed by server")
				log.Fatal(err)
			}

			if err == io.EOF {
				// Connection closed by client
				break
			}
			if errors.Is(err, syscall.ECONNRESET) {
				// Connection reset by peer
				break
			}
			log.Fatal(err)
		}

		chunkRead := string(buffer[0:bytesRead])

		// Checks message length to ensure check if there is more chunks to be read
		mLen := strings.IndexAny(chunkRead, "mLen")
		if mLen != -1 {
			messageSize, err = strconv.Atoi(chunkRead[0:mLen])

			if err != nil {
				log.Fatal(err)
			}

			message += string(buffer[mLen+4 : bytesRead])
		} else {
			message += string(buffer[0:bytesRead])
		}

		// Full message read
		if len([]rune(message)) == messageSize {
			messageCh <- message
			message = ""
			messageSize = 0
			if isSingleMessage {
				break
			}
		}
	}
}
