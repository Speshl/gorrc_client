package models

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

const ClientAxesCount = 10

type ConnectReq struct {
	Key      string `json:"key"`
	Password string `json:"password"`
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

type Offer struct {
	Offer        webrtc.SessionDescription `json:"offer"`
	CarShortName string                    `json:"car_name"`
}

type ControlState struct {
	Axes      []float64 `json:"axes"`
	BitButton uint32    `json:"bit_buttons"`
	Buttons   []bool
}

type Hud struct {
	Lines []string `json:"lines"`
}
