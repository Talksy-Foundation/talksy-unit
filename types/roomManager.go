package types

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type Lobby interface {
	CreateRoom(id string)
	RemoveRoom(id string)
	AddUserToRoom(self_id string, room_id string, socket *websocket.Conn)
	RemoveUserFromRoom(self_id string, room_id string, socket *websocket.Conn)
	ShowSessions()
	ObtainEvent(message WsMessage, socket *websocket.Conn)
}

type RoomManager struct {
	Sessions map[string]*Room
}

// NewRoomManager creates a new RoomManager instance.
//
// Returns a pointer to the created RoomManager.
func NewRoomManager() *RoomManager {
	return &RoomManager{Sessions: map[string]*Room{}}
}

// ShowSessions returns the map of room sessions.
//
// No parameters.
// Returns a map of string keys to Room values.
func (RoomManager *RoomManager) ShowSessions() map[string]*Room {
	return RoomManager.Sessions
}

// CreateRoom adds a new room with the given ID to the RoomManager's sessions.
//
// Parameters:
// - id: string representing the ID of the new room.
func (RoomManager *RoomManager) CreateRoom(id string) {
	RoomManager.Sessions[id] = NewRoom(id)
}

// RemoveRoom removes a room from the RoomManager by ID.
//
// Parameters:
// - id: string representing the ID of the room to be removed.
func (RoomManager *RoomManager) RemoveRoom(id string) {
	delete(RoomManager.Sessions, id)
}

// AddUserToRoom adds a user to a room in the RoomManager.
//
// Parameters:
// - self_id: string representing the user ID.
// - room_id: string representing the room ID.
// - socket: *websocket.Conn for the user's socket connection.
func (RoomManager *RoomManager) AddUserToRoom(self_id string, room_id string, socket *websocket.Conn) {
	if _, ok := RoomManager.Sessions[room_id]; !ok {
		log.Println("New Room was created: ", room_id)
		RoomManager.CreateRoom(room_id)
	}
	if room, ok := RoomManager.Sessions[room_id]; ok {
		// Add Peer to Room
		room.AddPeer(newPeer(self_id))
		log.Println("Peer ", self_id, "was added to room ", room_id)
		if peer, ok := room.peers[self_id]; ok {
			// Set socket connection to Peer
			peer.SetSocket(socket)

			// Create Peer Connection
			conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
			if err != nil {
				log.Println("Failed to establish peer connection")
			}

			peer.SetPeerConnection(conn)
			log.Println("Peer connection was established")
			// Accept one audio and one video track incoming
			for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
				if _, err := peer.connection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
					Direction: webrtc.RTPTransceiverDirectionRecvonly,
				}); err != nil {
					log.Print(err)
					return
				}
			}

			// If PeerConnection is closed remove it from global list
			peer.connection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
				switch p {
				case webrtc.PeerConnectionStateFailed:
					if err := peer.connection.Close(); err != nil {
						log.Print(err)
					}
				case webrtc.PeerConnectionStateClosed:
					room.Signal()
				default:
				}
			})

			// When peer connection is getting the ICE -> send ICE to client
			peer.connection.OnICECandidate(func(i *webrtc.ICECandidate) {
				if i == nil {
					log.Println("ICEGatheringState: connected")
					return
				}
				log.Println("Ice: ", i)
				room.SendICE(i, self_id)
			})

			peer.connection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
				log.Println("Track added from peer: ", self_id)
				defer room.Signal()
				// Create a track to signal our incoming video to all peers
				trackLocal := room.AddTrack(t)
				defer room.RemoveTrack(trackLocal)
				defer log.Println("Track", trackLocal, "was removed")
				buf := make([]byte, 1500)
				for {
					i, _, err := t.Read(buf)
					if err != nil {
						return
					}

					if _, err = trackLocal.Write(buf[:i]); err != nil {
						return
					}
				}
			})
		}

	}
}

// RemoveUserFromRoom removes a user from a room if the user and room exist.
//
// Parameters:
// - self_id: string
// - room_id: string
func (RoomManager *RoomManager) RemoveUserFromRoom(self_id string, room_id string) {
	if room, ok := RoomManager.Sessions[room_id]; ok {
		if _, ok := room.peers[self_id]; ok {
			delete(room.peers, self_id)
		}
	}
}

// ObtainEvent handles different events from WebSocket messages and performs corresponding actions.
//
// Parameters:
// - message: WsMessage struct containing the event and data.
// - socket: *websocket.Conn for WebSocket connection.
func (RoomManager *RoomManager) ObtainEvent(message WsMessage, socket *websocket.Conn) {
	wsMessage := message
	switch wsMessage.Event {
	case "joinRoom":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id := m["self_id"].(string)
				room_id := m["room_id"].(string)
				RoomManager.AddUserToRoom(self_id, room_id, socket)
			}
		}()
	case "leaveRoom":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id := m["self_id"].(string)
				room_id := m["room_id"].(string)
				RoomManager.RemoveUserFromRoom(self_id, room_id)
			}
		}()
	case "offer":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id, _ := m["self_id"].(string)
				room_id, _ := m["room_id"].(string)
				offer2 := m["offer"].(map[string]any)
				if room, ok := RoomManager.Sessions[room_id]; ok {
					if peer, ok := room.peers[self_id]; ok {
						answer, err2 := peer.ReactOnOffer(offer2["sdp"].(string))
						if err2 != nil {
							log.Println(err2)
							return
						}
						room.SendAnswer(answer, self_id)
					}
				}
			}
		}()
	case "answer":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id, _ := m["self_id"].(string)
				room_id, _ := m["room_id"].(string)
				offer2 := m["answer"].(map[string]any)
				if room, ok := RoomManager.Sessions[room_id]; ok {
					if peer, ok := room.peers[self_id]; ok {
						err := peer.ReactOnAnswer(offer2["sdp"].(string))
						if err != nil {
							log.Println(err)
							return
						}
					}

				}
			}
		}()
	case "ice-candidate":
		go func() {
			m, ok := message.Data.(map[string]any)
			if ok {
				self_id, _ := m["self_id"].(string)
				room_id, _ := m["room_id"].(string)
				candidate := m["candidate"].(map[string]any)
				i_candidate := candidate["candidate"].(string)
				sdp_mid := candidate["sdpMid"].(string)
				sdp_m_line_index := uint16(candidate["sdpMLineIndex"].(float64))
				var username_fragment string
				if candidate["usernameFragment"] != nil {
					username_fragment = candidate["usernameFragment"].(string)
				} else {
					username_fragment = ""
				}
				init := webrtc.ICECandidateInit{
					Candidate:        i_candidate,
					SDPMid:           &sdp_mid,
					SDPMLineIndex:    &sdp_m_line_index,
					UsernameFragment: &username_fragment,
				}
				if room, ok := RoomManager.Sessions[room_id]; ok {
					if peer, ok := room.peers[self_id]; ok {
						if err := peer.connection.AddICECandidate(init); err != nil {
							log.Println(err)
							return
						}
						log.Println("ICE-CANDIDATE added for peer", peer.id)
						log.Println(peer.connection.ICEConnectionState())
						log.Println(peer.connection.ICEGatheringState())
					}
				}
			} else {
				log.Println(m)
				log.Println("nach")
			}
		}()
	default:
		log.Println("DEFAULT")
		log.Println(wsMessage)
	}
}
