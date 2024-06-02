package api

// import (
// 	"fmt"
// 	"net/http"

// 	_ "talksy_unit/web"

// 	"github.com/Talksy-Foundation/sfu"
// )

// func (roomManager RoomManagerWrapper) CreateRoom(w http.ResponseWriter, r *http.Request) {
// 	roomID := roomManager.CreateRoomID()
// 	roomName := r.URL.Query().Get("name")
// 	if roomName == "" {
// 		roomName = roomID
// 	}
// 	roomsOpts := sfu.DefaultRoomOptions()
// 	roomsOpts.Bitrates.InitialBandwidth = 1_000_000
// 	// roomsOpts.PLIInterval = 3 * time.Second
// 	roomManager.NewRoom(roomID, roomName, sfu.RoomTypeLocal, roomsOpts)
// 	fmt.Fprintf(w, "room_id: %s, room_name: %s", roomID, roomName)
// }

// func RetrieveStats(w http.ResponseWriter, r *http.Request) {
// 	statsHandler(w, r, defaultRoom)
// }
