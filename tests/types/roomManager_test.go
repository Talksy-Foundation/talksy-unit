package types_tests

import (
	"testing"

	types "talksy_unit/types"

	"github.com/gorilla/websocket"
)

func TestAddUserToRoom_UserNotInRoom(t *testing.T) {
	// Create a RoomManager instance
	roomManager := &types.RoomManager{
		Sessions: make(map[string]*types.Room),
	}

	// Create a user ID, room ID, and a socket
	selfID := "user1"
	roomID := "room1"
	socket := &websocket.Conn{} // Mock socket connection

	roomManager.AddUserToRoom(selfID, roomID, socket)

	// Add your assertions here based on the expected behavior
}

func TestAddUserToRoom_UserInRoom(t *testing.T) {
	// Create a RoomManager instance
	roomManager := &types.RoomManager{
		Sessions: make(map[string]*types.Room),
	}

	// Create a user ID, room ID, and a socket
	selfID := "user1"
	roomID := "room1"
	socket := &websocket.Conn{} // Mock socket connection

	// Create the room and add the user once
	roomManager.CreateRoom(roomID)
	roomManager.AddUserToRoom(selfID, roomID, socket)

	// Add your assertions here based on the expected behavior
}
