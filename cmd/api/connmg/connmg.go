package connmg

import (
	"fmt"
	"strings"
	"sync"

	"github.com/thiagodejesus/tcp-chat-example/internal/conn"

	"github.com/google/uuid"
)

var mu sync.Mutex

type ActiveConnectionsMap map[uuid.UUID]*conn.LiveConnection

type ActiveConnections struct {
	Connections ActiveConnectionsMap
}

func (ac *ActiveConnections) BroadcastMessage(message string, origin conn.LiveConnection) {
	for _, value := range ac.Connections {
		value.WriteMessage(fmt.Sprintf("[%v] %v", origin.Id.String(), message), conn.ServerConn)
	}
}

func HandleNewConnection(liveConn *conn.LiveConnection, activeConnections *ActiveConnections) {
	liveConn.WriteMessage("id:"+liveConn.Id.String(), conn.ServerConn)

	mu.Lock()
	activeConnections.Connections[liveConn.Id] = liveConn
	mu.Unlock()
	fmt.Printf("Established connection [%v]\nActiveConnections: %v\n", liveConn.Id.String(), activeConnections.Connections)

	activeConnections.BroadcastMessage("Joined the room", *liveConn)

	messageCh := make(chan string)
	go conn.HandleReceivedMessages(messageCh, liveConn.Connection, false, conn.ServerConn)

	for message := range messageCh {
		fmt.Printf("Id: %v, has send message: %v", liveConn.Id.String(), message)
		if strings.HasPrefix(message, "\\all ") {
			activeConnections.BroadcastMessage(message[5:], *liveConn)
		}
		if strings.HasPrefix(message, "\\list") {
			activeIds := "Connected Ids:"

			for k := range activeConnections.Connections {
				activeIds += fmt.Sprintf("\n %v", k)
			}
			liveConn.WriteMessage(activeIds, conn.ServerConn)
		}
	}

	// Shut down the connection.
	fmt.Printf("\n[%v]Closing connection\n", liveConn.Id.String())
	liveConn.Connection.Close()

	mu.Lock()
	delete(activeConnections.Connections, liveConn.Id)
	mu.Unlock()
	fmt.Printf("Closed connection [%v]\nActiveConnections: %v\n", liveConn.Id.String(), activeConnections.Connections)
}
