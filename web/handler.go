package web

import (
	"time"

	"github.com/Talksy-Foundation/sfu"
	"github.com/Talksy-Foundation/sfu/pkg/interceptors/voiceactivedetector"
)

type RoomManagerWrapper struct {
	RoomManager *sfu.Manager
	DefaultRoom *sfu.Room
}

type APIError struct {
	ErrorCode    int
	ErrorMessage string
	CreatedAt    time.Time
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
