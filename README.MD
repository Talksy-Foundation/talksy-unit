# Talksy Unit

> [!TIP]  
> Single Forwarding Unit for Talksy conferencing service.

> [!IMPORTANT]
> Early development stage can mix SFU and MCU  

Talksy Unit is a server for transmitting video and audio in real time using
WebRTC technologies. This SFU (Selective Forwarding Unit) is designed to provide
high-quality video and audio conferencing, providing efficient and
scalable solution.


## Table Of Contents
1. [Structure](#structure)
    - [Peer](#peer)
    - [Room](#room)
    - [RoomManager](#roommanager)
2. [Features](#features)
   - [Trickle ICE](#trickle-ice) 
   - [Renegotiation](#re-negotiation)
3. [Setup](#setup)

## Structure

### Peer

> [!NOTE]  
> The Peer structure is an implementation of the PeerInterface interface and is used
> to manage the WebRTC connection, process offers and answers, and
> adding and deleting remote media tracks. Each Peer instance is associated with a unique identifier (id)
> and provides secure multithreaded communication using a mutex.

``` go
type PeerInterface interface {
    SetSocket(ws_conn *websocket.Conn)
    AddRemoteTrack(track *webrtc.TrackRemote)
    RemoveRemoteTrack(track *webrtc.TrackRemote)
    SetPeerConnection(conn *webrtc.PeerConnection)
    ReactOnOffer(offer webrtc.SessionDescription)
}
```
``` go
type Peer struct {
	id         string
	connection *webrtc.PeerConnection
	streams    map[string]*webrtc.TrackRemote
	mutex      sync.RWMutex
	socket     *websocket.Conn
}
```

### Room

> [!NOTE] 
> The Room structure represents the room object that manages the connection and media tracks
> for participants. It implements the Session interface and provides methods for adding and
> deleting participants, as well as processing media tracks. The Signal method is used for synchronization
> state of the room with the state of all participants.

``` go
 type Session interface {
	JoinRoom(id string)
	AddPeer(peer *Peer)
	RemovePeer(peer_id string)
	AddTrack(track *webrtc.TrackRemote)
	RemoveTrack(track *webrtc.TrackRemote)
	SendAnswer(message webrtc.SessionDescription, peer_id string)
	Signal()
}

```
``` go
type Room struct {
	id     string
	mutex  sync.RWMutex
	peers  map[string]*Peer
	tracks map[string]*webrtc.TrackLocalStaticRTP
} 
```

### RoomManager

> [!NOTE]   
> The RoomManager structure is a coordinator for managing rooms (Room). Implements
> Lobby interface and provides methods for creating, deleting rooms, adding and removing users
> from the room. The ObtainEvent method is used to handle events from WebSocket such as attach
> to the room, disconnecting, transmitting proposals and responses, and ICE candidates.

``` go
type Lobby interface {
	CreateRoom(id string)
	RemoveRoom(id string)
	AddUserToRoom(self_id string, room_id string, socket *websocket.Conn)
	RemoveUserFromRoom(self_id string, room_id string, socket *websocket.Conn)
}
```
``` go
type RoomManager struct {
	sessioins map[string]*Room
}
```

## Features
### Trickle ICE
Trickle ICE is the process of sharing addresses as soon as they are gathered. This parallelizes establishing a connection with a remote peer and starting sessions with TURN servers. Using Trickle ICE can dramatically reduce the amount of time it takes to establish a WebRTC connection.

> [!CAUTION] 
> Trickle ICE isn't mandatory to use, but highly recommended.

![ICE](https://i.vimeocdn.com/video/611350196-56fc3656cd8c8fa580d154012b92e26edb9e7fcda4fcd53f9cc4ce04d1a22adf-d)

### Re-negotiation

> [!TIP]  
> Renegotiation in the Twenties WebRTC Represents Process Changes
> connection parameters between participants in a communication session. This may include
> includes adding or removing media streams, changing codecs, as well as
> security options such as the use of encryption.

The renegotiation process is usually initiated by one of the session participants,
who wants to make changes to the present connection. This can lead to
for various reasons such as adding a new component, changing
quality transfer or connection of a new device.

## Setup
```bash
docker build -t talksy-unit:1.0.0
docker run -p 8080:8080 -d talksy-unit:1.0.0
```

### Testing:

* https://webrtc.github.io/samples/src/content/peerconnection/trickle-ice/

### References:

* https://webrtcforthecurious.com/