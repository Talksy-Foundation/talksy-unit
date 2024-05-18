# Talksy Unit

[![Go](https://github.com/Talksy-Foundation/talksy-unit/actions/workflows/go.yml/badge.svg)](https://github.com/Talksy-Foundation/talksy-unit/actions/workflows/go.yml)

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


```go
type Request struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Respose struct {
	Status bool        `json:"status"`
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
}

type VAD struct {
	SSRC     uint32                                `json:"ssrc"`
	TrackID  string                                `json:"track_id"`
	StreamID string                                `json:"stream_id"`
	Packets  []voiceactivedetector.VoicePacketData `json:"packets"`
}

type AvailableTrack struct {
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
	TrackID    string `json:"track_id"`
	StreamID   string `json:"stream_id"`
	Source     string `json:"source"`
}
```

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

### GitHub Personal Access Token
In order for Go to access the private repository, a Personal Access Token from GitHub will be needed. You can create an access token from the GitHub settings: https://github.com/settings/tokens. Give the token a description and select the “repo” scope.

### Configure Go with Personal Access Token
Once you have your access token, you’ll need to configure Go to use it when trying to access the private repository. You can do this by setting the GOPRIVATE environment variable to the hostname of the repository, and GONOPROXY variable to localhost. Finally set the global git config to use the token.

```sh
export GOPRIVATE=github.com/Talksy-Foundation/sfu
export GONOPROXY=localhost
export GITHUB_ACCESS_TOKEN=<your-token>

$Env:GOPRIVATE="github.com/Talksy-Foundation/sfu"
$Env:GONOPROXY="localhost"
$Env:GITHUB_ACCESS_TOKEN="<your-token>"

git config --global url."https://$GITHUB_ACCESS_TOKEN:x-oauth-basic@github.com/".insteadOf "https://github.com/"

go get github.com/Talksy-Foundation/sfu

# Paste <your-token>

```

```bash
$Env:TURN_ENABLED="TRUE"
go run talksyunit.go
```

1. Open the browser `http://localhost:8000/?room_id=1111` to open the client.
2. Click the Start button to start the WebRTC connection.
3. Open another browser tab `http://localhost:8000/?room_id=1111` to start a different client.
4. Click the Start button to start the WebRTC connection, you should see the video stream from the first client.
5. Repeat the steps above to add more clients.
6. Change `room_id` to create new room

```bash
docker build -t talksy-unit:1.0.0
docker run -p 8080:8000 -d talksy-unit:1.0.0
```

### Testing:

* https://webrtc.github.io/samples/src/content/peerconnection/trickle-ice/

### References:

* https://webrtcforthecurious.com/
* https://github.com/inlivedev/sfu/blob/main/examples/http-websocket/main.go
