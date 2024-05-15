package ws_server

import (
	"encoding/json"
	"log"
	"net/http"
	"talksy_unit/types"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},

	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WsServer struct {
	clients     map[*websocket.Conn]bool
	roomManager types.RoomManager
}

// StartServer initializes the WebSocket server.
//
// No parameters.
// Returns a pointer to the WsServer.
func StartServer() *WsServer {
	server := WsServer{
		make(map[*websocket.Conn]bool),
		*types.NewRoomManager(),
	}
	http.HandleFunc("/", server.wsInit)
	log.Println("Server started successfully")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Println(err)
	}

	return &server
}

// wsInit initializes the websocket connection.
//
// It takes a http.ResponseWriter and a http.Request as parameters.
func (ws *WsServer) wsInit(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)

	defer conn.Close()

	log.Printf("Client connected")

	if err != nil {
		log.Printf(" with error %s", err)
		return
	}

	log.Println(" successfully")

	message := types.WsMessage{}

	for {
		messageType, bmessage, err := conn.ReadMessage()

		if err != nil {
			log.Println(err)
			return
		}
		if messageType == websocket.CloseMessage {
			break
		}

		err = json.Unmarshal(bmessage, &message)
		if err != nil {
			log.Println("DROP")
			log.Println(message.Data)
			log.Println(err)
			return
		}
		ws.roomManager.ObtainEvent(message, conn)
	}
}
