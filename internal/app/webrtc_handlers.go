package app

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/pion/webrtc/v3"
)

const PingSourceName = "car"

func (c *Connection) onICEConnectionStateChange(connectionState webrtc.ICEConnectionState) {
	log.Printf("Connection State has changed: %s\n", connectionState.String())
}

func (c *Connection) onICECandidate(candidate *webrtc.ICECandidate) {
	if candidate != nil {
		log.Printf("recieved ICE candidate from client: %s\n", candidate.String())
	}
}

func (c *Connection) onDataChannel(d *webrtc.DataChannel) {
	log.Printf("new data channel: %s\n", d.Label())

	// Register channel opening handler
	d.OnOpen(func() {
		log.Printf("data channel open: %s\n", d.Label())
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
		log.Printf("recieved message on unsupported channel: %s\n", d.Label())
	}
}

func (c *Connection) onCommandHandler(data []byte) {
	state := models.ControlState{}
	err := json.Unmarshal(data, &state)
	if err != nil {
		log.Printf("failed unmarshalling data channel msg: %s\n", data)
		return
	}
	//log.Println("command recieved")
	c.CommandChannel <- state
}

func (c *Connection) onPingHandler(data []byte) {
	ping := models.Ping{}
	err := json.Unmarshal(data, &ping)
	if err != nil {
		log.Printf("failed unmarshalling data channel msg: %s\n", data)
		return
	}
	if ping.Source == PingSourceName {
		roundTripTime := time.Now().UnixMilli() - ping.TimeStamp
		log.Printf("ping: %d ms\n", roundTripTime)
	}
}
