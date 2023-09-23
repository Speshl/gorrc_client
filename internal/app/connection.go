package app

import (
	"context"
	"fmt"
	"log"

	"github.com/Speshl/gorrc_client/internal/models"
	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
)

type CommandHandler func(models.ControlState)

type Connection struct {
	// ID             string
	Socket         socketio.Conn
	PeerConnection *webrtc.PeerConnection
	Ctx            context.Context
	CtxCancel      context.CancelFunc
	CommandChannel chan models.ControlState
	HudChannel     chan models.Hud

	HudOutput *webrtc.DataChannel
}

func NewConnection(ctx context.Context, socketConn socketio.Conn, commandChan chan models.ControlState, hudChan chan models.Hud) (*Connection, error) {
	log.Printf("Creating User Connection %s\n", socketConn.ID())
	webrtcCfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(webrtcCfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Peer Connection: %s", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	conn := &Connection{
		// ID:             socketConn.ID(),
		Socket:         socketConn,
		PeerConnection: peerConnection,
		Ctx:            ctx,
		CtxCancel:      cancel,
		CommandChannel: commandChan,
		HudChannel:     hudChan,
	}
	return conn, nil
}

func (c *Connection) Disconnect() {
	c.CtxCancel()
	c.PeerConnection.Close()
}

func (c *Connection) RegisterHandlers(audioTrack *webrtc.TrackLocalStaticSample, videoTrack *webrtc.TrackLocalStaticSample) error {

	// _, err := c.PeerConnection.AddTrack(audioTrack)
	// if err != nil {
	// 	return fmt.Errorf("error adding audio track: %w", err)
	// }

	// _, err = c.PeerConnection.AddTrack(videoTrack)
	// if err != nil {
	// 	return fmt.Errorf("error adding video track: %w", err)
	// }

	//c.PeerConnection.OnTrack(c.AudioPlayer) //TODO: Uncomment to play client audio

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	c.PeerConnection.OnICEConnectionStateChange(c.onICEConnectionStateChange)

	// Handle ICE candidate messages from the client
	c.PeerConnection.OnICECandidate(c.onICECandidate)

	c.PeerConnection.OnDataChannel(c.onDataChannel)

	go func() {
		for {
			select {
			case <-c.Ctx.Done():
				log.Printf("stopping safety monitor: %s\n", c.Ctx.Err().Error())
				return
			case hud, ok := <-c.HudChannel:
				if !ok {
					log.Println("hud channel closed")
					return
				}
				if c.HudOutput != nil {
					encodedHud, err := encode(hud)
					err = c.HudOutput.SendText(encodedHud)
					if err != nil {
						log.Printf("failed sending hud: error - %s\n", err.Error())
						continue
					}
				}
			}
		}
	}()
	return nil
}
