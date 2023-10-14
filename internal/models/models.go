package models

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

const ClientAxesCount = 10

type ConnectReq struct {
	Key       string `json:"key"`
	Password  string `json:"password"`
	SeatCount int    `json:"seat_count"`
}

type ConnectResp struct {
	Car   Car
	Track Track
}
type Car struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
	Type      string    `json:"type"`
}

type Track struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
	Type      string    `json:"type"`
}

type IceCandidate struct {
	Candidate    webrtc.ICECandidateInit `json:"candidate"`
	CarShortName string                  `json:"car_name"`
	SeatNum      int                     `json:"seat_number"`
	UserId       uuid.UUID               `json:"user_id"`
}

type Offer struct {
	Offer        webrtc.SessionDescription `json:"offer"`
	CarShortName string                    `json:"car_name"`
	SeatNumber   int                       `json:"seat_number"`
	UserId       uuid.UUID                 `json:"user_id"`
}

type Answer struct {
	Answer     *webrtc.SessionDescription `json:"answer"`
	SeatNumber int                        `json:"seat_number"`
}

type ControlState struct {
	Axes      []float64 `json:"axes"`
	BitButton uint32    `json:"bit_buttons"`
	TimeStamp int64     `json:"time_stamp"`
	Buttons   []bool
}

type Hud struct {
	Lines []string `json:"lines"`
}

type Ping struct {
	Source    string `json:"source"`
	TimeStamp int64  `json:"time_stamp"`
}

type Seat struct {
	Index          int
	CommandChannel chan ControlState
	HudChannel     chan Hud
	VideoTracks    []*webrtc.TrackLocalStaticSample
	AudioTracks    []*webrtc.TrackLocalStaticSample
}
