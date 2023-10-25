package app

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/pion/webrtc/v3"
)

const PingSourceName = "car"
const PingWarningThreshold = 100 * time.Millisecond

func (c *Connection) onICEConnectionStateChange(connectionState webrtc.ICEConnectionState) {
	log.Printf("Connection State has changed: %s\n", connectionState.String())

	switch connectionState {
	case webrtc.ICEConnectionStateFailed:
		c.Disconnect()
	case webrtc.ICEConnectionStateChecking:
	case webrtc.ICEConnectionStateCompleted:
	case webrtc.ICEConnectionStateConnected:
	case webrtc.ICEConnectionStateClosed:
		c.Disconnect()
	case webrtc.ICEConnectionStateDisconnected:
	case webrtc.ICEConnectionStateNew:
	default:
	}
}

func (c *Connection) onICECandidate(candidate *webrtc.ICECandidate) {
	if candidate != nil {
		log.Printf("recieved ICE candidate from client: %s (not used)\n", candidate.String())
	}
}

func (c *Connection) onDataChannel(d *webrtc.DataChannel) {
	log.Printf("new data channel for seat %d: %s\n", c.SeatNumber, d.Label())

	// Register channel opening handler
	d.OnOpen(func() {
		log.Printf("data channel open for seat %d: %s\n", c.SeatNumber, d.Label())
		switch d.Label() {
		case "hud":
			c.HudOutput = d
		case "ping":
			c.PingOutput = d
		}
	})

	// Register text message handling
	switch d.Label() {
	case "command":
		d.OnMessage(func(msg webrtc.DataChannelMessage) { c.onCommandHandler(msg.Data) })
	case "ping":
		d.OnMessage(func(msg webrtc.DataChannelMessage) { c.onPingHandler(msg.Data) })
	case "hud":
	default:
		log.Printf("recieved message on unsupported channel for seat %d: %s\n", c.SeatNumber, d.Label())
	}
}

func (c *Connection) onCommandHandler(data []byte) {
	state := models.ControlState{}
	err := json.Unmarshal(data, &state)
	if err != nil {
		log.Printf("error: failed unmarshalling data channel msg: %s\n", data)
		return
	}
	c.CommandChannel <- state
}

func (c *Connection) onPingHandler(data []byte) {
	ping := models.Ping{}
	err := json.Unmarshal(data, &ping)
	if err != nil {
		log.Printf("error: failed unmarshalling data channel msg: %s\n", data)
		return
	}
	if ping.Source == PingSourceName {
		roundTripTime := time.Now().UnixMilli() - ping.TimeStamp
		if roundTripTime > PingWarningThreshold.Milliseconds() {
			log.Printf("warning: user ping > %d for seat %d: %d ms\n", PingWarningThreshold.Microseconds(), c.SeatNumber, roundTripTime)
		}
		c.PingInput <- roundTripTime
	}
}
