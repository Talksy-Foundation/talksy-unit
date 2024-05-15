package types

import (
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"

	"github.com/gorilla/websocket"
)

type PeerInterface interface {
	SetSocket(ws_conn *websocket.Conn)
	AddRemoteTrack(track *webrtc.TrackRemote)
	RemoveRemoteTrack(track *webrtc.TrackRemote)
	SetPeerConnection(conn *webrtc.PeerConnection)
	ReactOnOffer(offer webrtc.SessionDescription)
}

type Peer struct {
	id         string
	connection *webrtc.PeerConnection
	streams    map[string]*webrtc.TrackRemote
	mutex      sync.RWMutex
	socket     *websocket.Conn
}

// newPeer creates a new Peer instance with the given id.
//
// Parameters:
//
//	id: The unique identifier for the Peer.
//
// Return types:
//
//	*Peer: The newly created Peer instance.
func newPeer(id string) *Peer {
	return &Peer{id: id, mutex: sync.RWMutex{}}
}

// SetPeerConnection sets the WebRTC PeerConnection for the Peer instance.
//
// Parameters:
//
//	conn: The WebRTC PeerConnection to be set.
//
// Return types:
func (peer *Peer) SetPeerConnection(conn *webrtc.PeerConnection) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	peer.connection = conn
}

// AddRemoteTrack adds a remote track to the Peer instance.
//
// Parameters:
//
//	track: The remote track to be added.
func (peer *Peer) AddRemoteTrack(track *webrtc.TrackRemote) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	peer.streams[track.ID()] = track
}

// RemoveRemoteTrack removes a remote track from the Peer instance.
//
// Parameters:
//
//	track: The remote track to be removed.
//
// Return types:
func (peer *Peer) RemoveRemoteTrack(track *webrtc.TrackRemote) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	delete(peer.streams, track.ID())
}

// SetSocket sets the socket for the Peer instance.
//
// Parameters:
//
//	socket: The socket to be set.
func (peer *Peer) SetSocket(socket *websocket.Conn) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	peer.socket = socket
}

// ReactOnOffer reacts to the offer received from a peer by setting remote and local descriptions.
//
// Parameters:
//
//	offer_str: The offer string received from the peer.
//
// Return types:
//
//	webrtc.SessionDescription: The session description for the answer.
//	error: An error if any occurred during the process.
func (peer *Peer) ReactOnOffer(offer_str string) (webrtc.SessionDescription, error) {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer_str,
	}
	err := peer.connection.SetRemoteDescription(offer)
	if err != nil {
		fmt.Println(err)
		return offer, err
	}
	log.Println("Remote Description was set in peer ", peer.id)
	answer, err := peer.connection.CreateAnswer(nil)
	_ = peer.connection.SetLocalDescription(answer)
	log.Println("Local Description was set for peer ", peer.id)
	if err != nil {
		return offer, err
	}
	log.Println("Answer was created in peer ", peer.id)
	return answer, nil

}

// ReactOnAnswer reacts to the answer received from a peer by setting the remote description.
//
// Parameters:
//
//	answer_str: The answer string received from the peer.
//
// Return types:
//
//	error: An error if any occurred during the process.
func (peer *Peer) ReactOnAnswer(answer_str string) error {
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answer_str,
	}
	err := peer.connection.SetRemoteDescription(answer)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
