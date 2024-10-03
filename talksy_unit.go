package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Talksy-Foundation/sfu"
	"github.com/Talksy-Foundation/sfu/pkg/fakeclient"
	"github.com/Talksy-Foundation/sfu/pkg/interceptors/voiceactivedetector"
	"github.com/Talksy-Foundation/sfu/pkg/networkmonitor"
	"github.com/golang/glog"
	"github.com/pion/logging"
	"github.com/pion/webrtc/v4"
	"golang.org/x/net/websocket"
)

type WebServer interface {
	statsHandler(w http.ResponseWriter, r *http.Request, room *sfu.Room)
	reader(conn *websocket.Conn, messageChan chan Request)
	clientHandler(isDebug bool, conn *websocket.Conn, messageChan chan Request, r *sfu.Room)
}

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

const (
	TypeOffer                = "offer"
	TypeAnswer               = "answer"
	TypeCandidate            = "candidate"
	TypeNetworkCondition     = "network_condition"
	TypeError                = "error"
	TypeAllowRenegotiation   = "allow_renegotiation"
	TypeIsAllowRenegotiation = "is_allow_renegotiation"
	TypeTrackAdded           = "tracks_added"
	TypeTracksAvailable      = "tracks_available"
	TypeSubscribeTracks      = "subscribe_tracks"
	TypeSwitchQuality        = "switch_quality"
	TypeUpdateBandwidth      = "update_bandwidth"
	TypeSetBandwidthLimit    = "set_bandwidth_limit"
	TypeBitrateAdjusted      = "bitrate_adjusted"
	TypeTrackStats           = "track_stats"
	TypeVoiceDetected        = "voice_detected"
)

var (
	RoomManager *sfu.Manager
	DefaultRoom *sfu.Room
)

//	@title			Swagger SFU Unit API
//	@version		1.0
//	@description	This is a SFU Unit.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	support@talksy.tuchacloud.ru
//	@contact.url	https://talksy.tuchacloud.ru/support
//	@contact.email	support@talksy.tuchacloud.ru

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		mu-1.talksy.tuchacloud.ru
//	@BasePath	/v2

// main starts the server using the custom wrapper.
//
// No parameters.

func main() {
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	ipAddr := GetOutboundIP().String()
	if _, ok := os.LookupEnv("DEBUG"); ok {
		ipAddr = "localhost"
	}

	port := ":8080"
	if p, ok := os.LookupEnv("PORT"); ok {
		port = ":" + p
	}

	sfuOpts := sfu.DefaultOptions()
	// sfuOpts.EnableMux = true
	sfuOpts.EnableBandwidthEstimator = true

	_, turnEnabled := os.LookupEnv("TURN_ENABLED")
	if turnEnabled {
		sfu.StartTurnServer(ctx, "127.0.0.1")
		// sfu.StartTurnServer(ctx, ipAddr)
		sfuOpts.IceServers = append(sfuOpts.IceServers, webrtc.ICEServer{
			URLs: []string{"turn:127.0.0.1:3478"},
			// URLs:           []string{"turn:" + ipAddr + ":3478"},
			Username:       "user",
			Credential:     "pass",
			CredentialType: webrtc.ICECredentialTypePassword,
		})
		log.Printf("Turn Server on http://%s:3478 ...", ipAddr)

	}

	localIp, _ := sfu.GetLocalIp()
	iceServers := []webrtc.ICEServer{
		{
			URLs: []string{
				"stun:stun.l.google.com:19302",
				"stun:stun1.l.google.com:19302",
				"stun:stun2.l.google.com:19302",
				"stun:stun3.l.google.com:19302",
				"stun:stun4.l.google.com:19302",
				// "stun:coturn.srv1.tuchacloud.ru:5349",
			},
		},
		{
			URLs: []string{
				"turn:" + localIp.String() + ":3478",
				"stun:" + localIp.String() + ":3478",
				"turn:" + ipAddr + ":3478",
				"stun:" + ipAddr + ":3478",
			},
			Username:       "user",
			Credential:     "pass",
			CredentialType: webrtc.ICECredentialTypePassword,
		},
		// {
		// 	URLs:           []string{"turn:coturn.srv1.tuchacloud.ru:3478?transport=udp"},
		// 	Username:       "test123",
		// 	Credential:     "test",
		// 	CredentialType: webrtc.ICECredentialTypePassword,
		// },
	}

	// create room manager first before create new room
	RoomManager = sfu.NewManager(ctx, "talksy-main-server", sfuOpts)

	DefaultRoom, _ = RoomManager.NewRoom("default", "default", sfu.RoomTypeLocal, sfu.DefaultRoomOptions())

	// multiple room can be created by calling this API endpoint
	http.HandleFunc("/create_room", func(w http.ResponseWriter, r *http.Request) {
		roomID := RoomManager.CreateRoomID()
		roomName := r.URL.Query().Get("name")
		if roomName == "" {
			roomName = roomID
		}
		roomsOpts := sfu.DefaultRoomOptions()
		roomsOpts.Bitrates.InitialBandwidth = 1_000_000
		// roomsOpts.PLIInterval = 3 * time.Second
		RoomManager.NewRoom(roomID, roomName, sfu.RoomTypeLocal, roomsOpts)
		fmt.Fprintf(w, "room_id: %s, room_name: %s", roomID, roomName)
	})

	fakeClientCount := 0

	for i := 0; i < fakeClientCount; i++ {
		log := logging.LeveledLogger(&logging.DefaultLeveledLogger{})
		log.Debug("fake-client")

		// WARN: panic: no video file
		fc := fakeclient.Create(ctx, log, DefaultRoom, iceServers, fmt.Sprintf("fake-client-%d", i), true) // create a fake client

		fc.Client.OnTracksAdded(func(addedTracks []sfu.ITrack) {
			setTracks := make(map[string]sfu.TrackType, 0)
			for _, track := range addedTracks {
				setTracks[track.ID()] = sfu.TrackTypeMedia
			}
			fc.Client.SetTracksSourceType(setTracks)
		})
	}

	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)

	http.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
		messageChan := make(chan Request)
		isDebug := false
		if conn.Request().URL.Query().Get("debug") != "" {
			isDebug = true
		}
		if conn.Request().URL.Query().Get("room_id") != "" {
			roomID := conn.Request().URL.Query().Get("room_id")
			room, _ := RoomManager.GetRoom(roomID)
			if room == nil {
				room, _ = RoomManager.NewRoom(roomID, roomID, sfu.RoomTypeLocal, sfu.DefaultRoomOptions())
			}
			go clientHandler(isDebug, conn, messageChan, room)
		} else {
			go clientHandler(isDebug, conn, messageChan, DefaultRoom)
		}

		reader(conn, messageChan)
	}))

	http.Handle("/wss", websocket.Handler(func(conn *websocket.Conn) {
		messageChan := make(chan Request)
		isDebug := false
		if conn.Request().URL.Query().Get("debug") != "" {
			isDebug = true
		}
		if conn.Request().URL.Query().Get("room_id") != "" {
			roomID := conn.Request().URL.Query().Get("room_id")
			room, _ := RoomManager.GetRoom(roomID)
			if room == nil {
				room, _ = RoomManager.NewRoom(roomID, roomID, sfu.RoomTypeLocal, sfu.DefaultRoomOptions())
			}
			go clientHandler(isDebug, conn, messageChan, room)
		} else {
			go clientHandler(isDebug, conn, messageChan, DefaultRoom)
		}

		reader(conn, messageChan)
	}))

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		statsHandler(w, r, DefaultRoom)
	})

	switch runtime.GOOS {
	case "linux":
		go ListenTCP(ipAddr, port)
		go ListenTLS(ipAddr, port)

	case "windows":
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Panic(err)
		}
	default:
		log.Panicf("Unsupported OS: %s", runtime.GOOS)
	}
}

func ListenTCP(ipAddr string, port string) {
	srv := &http.Server{}
	tcpL, err := net.Listen("tcp4", ipAddr+port)
	log.Printf("Listening on http://%s ...", ipAddr+port)
	srv.Serve(tcpL)
	if err != nil {
		log.Panic(err)
	}
}

func ListenTLS(ipAddr string, port string) {
	err := http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Printf("Listening on https://%s:443", ipAddr)
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// statsHandler handles the HTTP request for retrieving statistics of a room.
//
// It takes in the following parameters:
// - w: an http.ResponseWriter object used to write the response.
// - r: an *http.Request object representing the incoming request.
// - room: a pointer to an sfu.Room object representing the room for which statistics are requested.
//
// This function does not return anything.
func statsHandler(w http.ResponseWriter, r *http.Request, room *sfu.Room) {
	stats := room.Stats()

	statsJSON, _ := json.Marshal(stats)

	w.Header().Set("Content-Type", "application/json")

	_, _ = w.Write([]byte(statsJSON))
}

// reader is a function that reads messages from a WebSocket connection and sends them to a message channel.
//
// It takes two parameters:
// - conn: a pointer to a websocket.Conn object representing the WebSocket connection.
// - messageChan: a channel of type Request to which the received messages will be sent.
//
// The function runs in a loop until the context is cancelled. It uses a nested loop to continuously read messages from the connection.
// If an error occurs while decoding a message, it logs the error and continues to the next message.
// The received messages are sent to the messageChan channel.
//
// The function does not return any values.
func reader(conn *websocket.Conn, messageChan chan Request) {
	ctx, cancel := context.WithCancel(conn.Request().Context())
	defer cancel()

MessageLoop:
	for {
		select {
		case <-ctx.Done():
			break MessageLoop
		default:
			for {
				decoder := json.NewDecoder(conn)
				var req Request
				err := decoder.Decode(&req)
				if err != nil {
					if err.Error() == "EOF" {
						continue
					}

					glog.Info("error decoding message", err)
				}
				messageChan <- req
			}
		}
	}
}

// clientHandler handles a client connection and manages the client's lifecycle.
//
// Parameters:
// - isDebug: a boolean indicating whether debug mode is enabled.
// - conn: a pointer to a websocket connection.
// - messageChan: a channel of type Request used for sending requests to the client.
// - r: a pointer to an sfu.Room object representing the room the client is connected to.
func clientHandler(isDebug bool, conn *websocket.Conn, messageChan chan Request, r *sfu.Room) {
	ctx, cancel := context.WithCancel(conn.Request().Context())
	defer cancel()

	// create new client id, you can pass a unique int value to this function
	// or just use the SFU client counter
	clientID := r.CreateClientID()

	// add a new client to room
	// you can also get the client by using r.GetClient(clientID)
	opts := sfu.DefaultClientOptions()
	opts.EnableVoiceDetection = true
	opts.ReorderPackets = false
	client, err := r.AddClient(clientID, clientID, opts)
	if err != nil {
		log.Panic(err)
		return
	}

	if isDebug {
		client.EnableDebug()
	}

	defer r.StopClient(client.ID())

	_, _ = conn.Write([]byte("{\"type\":\"clientid\",\"data\":\"" + clientID + "\"}"))

	answerChan := make(chan webrtc.SessionDescription)

	// client.SubscribeAllTracks()

	client.OnTracksAdded(func(tracks []sfu.ITrack) {
		tracksAdded := map[string]map[string]string{}
		for _, track := range tracks {
			tracksAdded[track.ID()] = map[string]string{"id": track.ID()}
		}
		resp := Respose{
			Status: true,
			Type:   TypeTrackAdded,
			Data:   tracksAdded,
		}

		trackAddedResp, _ := json.Marshal(resp)

		_, _ = conn.Write(trackAddedResp)
	})

	client.OnTracksAvailable(func(tracks []sfu.ITrack) {
		if client.IsDebugEnabled() {
			glog.Info("tracks available", tracks)
		}
		tracksAvailable := map[string]map[string]interface{}{}
		for _, track := range tracks {

			tracksAvailable[track.ID()] = map[string]interface{}{
				"id":           track.ID(),
				"client_id":    track.ClientID(),
				"source_type":  track.SourceType().String(),
				"kind":         track.Kind().String(),
				"is_simulcast": track.IsSimulcast(),
			}
		}
		resp := Respose{
			Status: true,
			Type:   TypeTracksAvailable,
			Data:   tracksAvailable,
		}

		trackAddedResp, _ := json.Marshal(resp)

		_, _ = conn.Write(trackAddedResp)
	})

	client.OnRenegotiation(func(ctx context.Context, offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
		// SFU request a renegotiation, send the offer to client
		glog.Info("receive renegotiation offer from SFU")

		resp := Respose{
			Status: true,
			Type:   TypeOffer,
			Data:   offer,
		}

		sdpBytes, _ := json.Marshal(resp)

		_, _ = conn.Write(sdpBytes)

		// wait for answer from client
		ctxTimeout, cancelTimeout := context.WithTimeout(client.Context(), 30*time.Second)

		defer cancelTimeout()

		// this will wait for answer from client in 30 seconds or timeout
		select {
		case <-ctxTimeout.Done():
			glog.Error("timeout on renegotiation")
			return webrtc.SessionDescription{}, errors.New("timeout on renegotiation")
		case answer := <-answerChan:
			glog.Info("received answer from client ", client.Type(), client.ID())
			return answer, nil
		}
	})

	client.OnAllowedRemoteRenegotiation(func() {
		// SFU allow a remote renegotiation
		glog.Info("receive allow remote renegotiation from SFU")

		resp := Respose{
			Status: true,
			Type:   TypeAllowRenegotiation,
			Data:   "ok",
		}

		respBytes, _ := json.Marshal(resp)

		_, _ = conn.Write(respBytes)
	})

	client.OnIceCandidate(func(ctx context.Context, candidate *webrtc.ICECandidate) {
		// SFU send an ICE candidate to client
		resp := Respose{
			Status: true,
			Type:   TypeCandidate,
			Data:   candidate,
		}
		candidateBytes, _ := json.Marshal(resp)

		_, _ = conn.Write(candidateBytes)
	})

	client.OnNetworkConditionChanged(func(condition networkmonitor.NetworkConditionType) {
		resp := Respose{
			Status: true,
			Type:   TypeNetworkCondition,
			Data:   condition,
		}
		respBytes, _ := json.Marshal(resp)

		_, _ = conn.Write(respBytes)
	})

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := client.Stats()

			resp := Respose{
				Status: true,
				Type:   TypeTrackStats,
				Data:   stats,
			}

			respBytes, _ := json.Marshal(resp)
			_, _ = conn.Write(respBytes)

		case req := <-messageChan:
			// handle as SDP if no error
			if req.Type == TypeOffer || req.Type == TypeAnswer {
				var resp Respose

				sdp, _ := req.Data.(string)

				if req.Type == TypeOffer {
					// handle as offer SDP
					answer, err := client.Negotiate(webrtc.SessionDescription{SDP: sdp, Type: webrtc.SDPTypeOffer})
					if err != nil {
						glog.Error("error on negotiate", err)

						resp = Respose{
							Status: false,
							Type:   TypeError,
							Data:   err.Error(),
						}
					} else {
						// send the answer to client
						resp = Respose{
							Status: true,
							Type:   TypeAnswer,
							Data:   answer,
						}
					}

					respBytes, _ := json.Marshal(resp)

					conn.Write(respBytes)
				} else {
					glog.Info("receive renegotiation answer from client")
					// handle as answer SDP as part of renegotiation request from SFU
					// pass the answer to onRenegotiation handler above
					answerChan <- webrtc.SessionDescription{SDP: sdp, Type: webrtc.SDPTypeAnswer}
				}

				// don't continue execution
				continue
			} else if req.Type == TypeCandidate {
				candidate := webrtc.ICECandidateInit{
					Candidate: req.Data.(string),
				}
				err := client.AddICECandidate(candidate)
				if err != nil {
					log.Panic("error on add ice candidate", err)
				}
			} else if req.Type == TypeTrackAdded {
				setTracks := make(map[string]sfu.TrackType, 0)
				for id, trackType := range req.Data.(map[string]interface{}) {
					if trackType.(string) == "media" {
						setTracks[id] = sfu.TrackTypeMedia
					} else {
						setTracks[id] = sfu.TrackTypeScreen
					}
				}
				client.SetTracksSourceType(setTracks)
			} else if req.Type == TypeSubscribeTracks {
				subTracks := make([]sfu.SubscribeTrackRequest, 0)
				tracks, ok := req.Data.([]interface{})
				if ok {
					for _, track := range tracks {
						trackData, ok := track.(map[string]interface{})
						if ok {
							subTrack := sfu.SubscribeTrackRequest{
								ClientID: trackData["client_id"].(string),
								TrackID:  trackData["track_id"].(string),
							}

							subTracks = append(subTracks, subTrack)
						}
					}

					if err := client.SubscribeTracks(subTracks); err != nil {
						glog.Error("error on subscribe tracks", err)
					}
				} else {
					glog.Error("error on subscribe tracks wrong data format ", req.Data)
				}

			} else if req.Type == TypeSwitchQuality {
				quality := req.Data.(string)
				switch quality {
				case "low":
					log.Println("switch to low quality")
					client.SetQuality(sfu.QualityLow)
				case "mid":
					log.Println("switch to mid quality")
					client.SetQuality(sfu.QualityMid)
				case "high":
					log.Println("switch to high quality")
					client.SetQuality(sfu.QualityHigh)
				case "none":
					log.Println("switch to high quality")
					client.SetQuality(sfu.QualityNone)
				}
			} else if req.Type == TypeUpdateBandwidth {
				bandwidth := uint32(req.Data.(float64))
				client.UpdatePublisherBandwidth(bandwidth)
			} else if req.Type == TypeSetBandwidthLimit {
				bandwidth, _ := strconv.ParseUint(req.Data.(string), 10, 32)
				client.SetReceivingBandwidthLimit(uint32(bandwidth))
			} else if req.Type == TypeIsAllowRenegotiation {
				resp := Respose{
					Status: true,
					Type:   TypeAllowRenegotiation,
					Data:   client.IsAllowNegotiation(),
				}

				respBytes, _ := json.Marshal(resp)

				conn.Write(respBytes)

			} else {
				glog.Error("unknown message type", req)
			}
		}
	}
}
