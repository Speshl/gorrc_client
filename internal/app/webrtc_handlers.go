package app

import (
	"encoding/json"
	"log"

	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/pion/webrtc/v3"
)

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
		if d.Label() == "hud" {
			c.HudOutput = d
		}
	})

	// Register text message handling
	switch d.Label() {
	case "command":
		d.OnMessage(func(msg webrtc.DataChannelMessage) { c.onCommandHandler(msg.Data) })
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
	c.CommandChannel <- state
}
