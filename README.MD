# Talksy Unit

[![Go](https://github.com/Talksy-Foundation/talksy-unit/actions/workflows/go.yml/badge.svg)](https://github.com/Talksy-Foundation/talksy-unit/actions/workflows/go.yml)

![HiveLogo](.github/media/hive_logo.jpg)

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
   - [WebSerber](#webserver)
   - [Request](#request)
   - [Respose](#respose)
   - [VAD](#vad)
   - [AvailableTrack](#availabletrack)
2. [Features](#features)
   - [Trickle ICE](#trickle-ice)
   - [Renegotiation](#re-negotiation)
3. [Setup](#setup)

## Structure

### Webserver

The WebServer interface defines methods:

```go
type WebServer interface {
	statsHandler(w http.ResponseWriter, r *http.Request, room *sfu.Room)
	reader(conn *websocket.Conn, messageChan chan Request)
	clientHandler(isDebug bool, conn *websocket.Conn, messageChan chan Request, r *sfu.Room)
}
```

> [!NOTE]
>
> - `statsHandler(w http.ResponseWriter, r *http.Request, room *sfu.Room)`: This method handles the HTTP request for retrieving statistics of a room. It takes in the response writer, request, and a pointer to an `sfu.Room` object and writes the statistics as a JSON response.
> - `reader(conn *websocket.Conn, messageChan chan Request)`: This method reads messages from a WebSocket connection and sends them to a message channel. It takes a pointer to a `websocket.Conn` object and a channel of type `Request` as parameters. It runs in a loop until the context is cancelled and decodes incoming messages, sending them to the message channel.
> - `clientHandler(isDebug bool, conn *websocket.Conn, messageChan chan Request, r *sfu.Room)`: This method handles the WebSocket connection for a client. It takes a boolean flag for debugging, a pointer to a `websocket.Conn` object, a channel of type `Request`, and a pointer to an `sfu.Room` objectas parameters. It runs in a loop until the context is cancelled and handles incoming messages, performing actions based on the message type.

### Request

```go
type Request struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
```

> [!NOTE]
> The Request struct in Go defines a structure that represents a request.
> It has two fields: `Type` of type `string` and `Data` of type `interface{}`. The `Type` field is tagged with `json:"type"` to indicate that it should be serialized as "type" when the struct is marshaled to JSON. The `Data` field is also tagged with `json:"data"` to indicate that it should be serialized as "data". This struct is commonly used for representing requests in an API.

### Respose

```go
type Respose struct {
	Status bool        `json:"status"`
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
}
```

> [!NOTE]
> Respose is a Go struct that represents a response object. It has three fields:
>
> - `Status`: A boolean field that indicates the status of the response.
> - `Type`: A `string` field that represents the type of the response.
> - `Data`: An `interface{}` field that can hold any type of data. This field is typically used when the response data can be of various types.

### VAD

> [!NOTE]
> VAD is a data structure that represents a Voice Activity Detection (VAD) object. It contains the following fields:
>
> - `SSRC`: An unsigned 32-bit integer that represents a synchronization source identifier (SSRC). This is typically used in multimedia communication protocols like RTP (Real-time Transport Protocol).
> - `TrackID`: A `string` that represents the ID of a track. In the context of multimedia, a track can refer to an audio or video stream.
> - `StreamID`: A `string` that represents the ID of a stream. In the context of multimedia, a stream can refer to a collection of audio or video tracks.
> - `Packets`: A slice of `voiceactivedetector.VoicePacketData` objects. This suggests that the VAD struct is used to hold information about voice packets, likely used in the context of voice activity detection.

```go
type VAD struct {
	SSRC     uint32                                `json:"ssrc"`
	TrackID  string                                `json:"track_id"`
	StreamID string                                `json:"stream_id"`
	Packets  []voiceactivedetector.VoicePacketData `json:"packets"`
}
```

### AvailableTrack

> [!NOTE]
> AvailableTrack is a struct that represents a track that is available for a client in a system.

    It has the following fields:

> - `ClientID` (string): The ID of the client that owns the track.
> - `ClientName` (string): The name of the client that owns the track.
> - `TrackID` (string): The ID of the track.
> - `StreamID` (string): The ID of the stream that the track belongs to.
> - `Source` (string): The source of the track.

```go
type AvailableTrack struct {
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
	TrackID    string `json:"track_id"`
	StreamID   string `json:"stream_id"`
	Source     string `json:"source"`
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

> [!NOTE]
> In order for Go to access the private repository, a Personal Access Token from GitHub will be needed. You can create an access token from the GitHub settings: https://github.com/settings/tokens. Give the token a description and select the “repo” scope.

### Configure Go with Personal Access Token

> [!NOTE]
> Once you have your access token, you’ll need to configure Go to use it when trying to access the private repository. You can do this by setting the `GOPRIVATE` environment variable to the hostname of the repository, and `GONOPROXY` variable to localhost. Finally set the global git config to use the token.

### Development:

Unix:

```sh
export GOPRIVATE=github.com/Talksy-Foundation/sfu &&
export GONOPROXY=localhost &&
export GITHUB_ACCESS_TOKEN=<your-token>
```

Windows:

```sh
$Env:GOPRIVATE="github.com/Talksy-Foundation/sfu" 
$Env:GONOPROXY="localhost" 
$Env:GITHUB_ACCESS_TOKEN="<your-token>"
```

```sh
git config --global url."https://$GITHUB_ACCESS_TOKEN:x-oauth-basic@github.com/".insteadOf "https://github.com/"

go get github.com/Talksy-Foundation/sfu
# Paste <your-token> when password is requested
```

#### Environment Variables

| Name         | Default Value |
| ------------ | ------------- |
| TURN_ENABLED | TRUE          |
| DEBUG        | FALSE         |
| PORT         | 8080          |

Windows:

```bash
$Env:TURN_ENABLED="TRUE"
$Env:DEBUG="TRUE"
$Env:PORT=8080
go run talksy_unit.go
```

If you are using this for a development environment:

* Go to: `chrome://flags/#unsafely-treat-insecure-origin-as-secure`
* Enable `Insecure origins treated as secure`
* Add the addresses for which you want to ignore this policy
* Restart chrome

1. Open the browser `http://localhost:8080/?room_id=111&debug=true&dev=1` to open the client.
2. Click the Start button to start the WebRTC connection.
3. Open another browser tab `http://localhost:8080/?room_id=111&debug=true&dev=1` to start a different client.
4. Click the Start button to start the WebRTC connection, you should see the video stream from the first client.
5. Repeat the steps above to add more clients.
6. Change `room_id` to create new room

Docker:

```bash
docker build -t talksy-unit:1.0.0 .
docker run -p 8121:8080 -d talksy-unit:1.0.0
```

### Testing:

- https://webrtc.github.io/samples/src/content/peerconnection/trickle-ice/

<!-- ### Swagger UI:

```sh
swag init -g talks_yunit.go
swag fmt
``` -->

## Production:

### Build and Deploy

#### AMD64 (x86_64)

```sh
$Env:GOOS="linux"; $Env:GOARCH="amd64";
go build -o bin/talksy-unit-amd64-linux talksy_unit.go

scp ./bin/talksy-unit-amd64-linux USER@SERVER:/home/talksy-unit/talksy-unit-amd64-linux
scp -r ./bin USER@SERVER:/home/talksy-unit/
sudo chmod 755 talksy-unit-amd64-linux
export PORT=8080 TURN_ENABLED=FALSE ./talksy-unit-amd64-linux
```

#### ARM64

```sh
$Env:GOOS="linux"; $Env:GOARCH="arm64";
go build -o bin/talksy-unit-arm64-linux talksy_unit.go
sudo chmod 755 talksy-unit-arm64-linux
```

#### Run as system service:

Create a service unit file with a .service extension. This file defines the properties and behaviour of the background process. For example, let's create a file named myprocess.service:
`sudo nano /etc/systemd/system/talksy-unit.service`

In the service unit file, add the necessary configuration. Here's an example:

```bash
[Unit]
Description=Talksy Unit Process
After=network.target

[Service]
ExecStart=/home/talksy-unit/talksy-unit-amd64-linux
Type=simple
Restart=always

[Install]
WantedBy=default.target
```

Save the file and exit the text editor.
Reload systemd to read the new service unit file:
`sudo systemctl daemon-reload`

Start the background process using systemd:
`sudo systemctl start talksy-unit`

`sudo systemctl stop talksy-unit`

The background process will now start and run independently as a system service. You can view its status, stop it, or configure it to start automatically on system boot using various systemctl commands.

`sudo netstat -ntlp | grep :8080`

#### Docker:

```bash
sudo docker build -t talksy-unit:1.0.2 . && \
sudo docker run -p 8121:8080 -p 3478:3478/udp -p 3478:3478 -p 49152:65535/udp -p 49152:65535 -e TURN_ENABLED=TRUE -d talksy-unit:1.0.2
```

#### WSS

Use certificates to establish `wss` connection and enable `getMediaDevices` work

Testing:

* https://piehost.com/websocket-tester

### References:

- https://book.systemsapproach.org/foundation/requirements.html
- https://webrtcforthecurious.com/
- https://github.com/inlivedev/sfu/blob/main/examples/http-websocket/main.go
