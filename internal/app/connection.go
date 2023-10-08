package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Speshl/gorrc_client/internal/models"
	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
)

type AudioPlayer func(*webrtc.TrackRemote, *webrtc.RTPReceiver)

type CommandHandler func(models.ControlState)

type Connection struct {
	// ID             string
	Socket         socketio.Conn
	PeerConnection *webrtc.PeerConnection
	Ctx            context.Context
	CtxCancel      context.CancelFunc
	CommandChannel chan models.ControlState
	HudChannel     chan models.Hud

	Speaker AudioPlayer

	HudOutput  *webrtc.DataChannel
	PingOutput *webrtc.DataChannel
}

func NewConnection(socketConn socketio.Conn, commandChan chan models.ControlState, hudChan chan models.Hud, speakers AudioPlayer) (*Connection, error) {
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

	ctx, cancel := context.WithCancel(context.Background())
	conn := &Connection{
		// ID:             socketConn.ID(),
		Socket:         socketConn,
		PeerConnection: peerConnection,
		Ctx:            ctx,
		CtxCancel:      cancel,
		CommandChannel: commandChan,
		HudChannel:     hudChan,
		Speaker:        speakers,
	}
	return conn, nil
}

func (c *Connection) Disconnect() {
	log.Println("user disconnecting")
	c.CtxCancel()
	c.PeerConnection.Close()
}

func (c *Connection) RegisterHandlers(audioTrack []*webrtc.TrackLocalStaticSample, videoTrack []*webrtc.TrackLocalStaticSample) error {

	log.Println("adding audio track")
	_, err := c.PeerConnection.AddTrack(audioTrack[0]) //TODO add all audio tracks
	if err != nil {
		return fmt.Errorf("error adding audio track: %w", err)
	}

	log.Println("adding video track")
	_, err = c.PeerConnection.AddTrack(videoTrack[0]) //TODO add all video tracks
	if err != nil {
		return fmt.Errorf("error adding video track: %w", err)
	}

	log.Println("set client track player")
	c.PeerConnection.OnTrack(c.Speaker) //TODO: Update this to kick out video tracks

	log.Println("start event listeners")
	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	c.PeerConnection.OnICEConnectionStateChange(c.onICEConnectionStateChange)

	// Handle ICE candidate messages from the client
	c.PeerConnection.OnICECandidate(c.onICECandidate)

	c.PeerConnection.OnDataChannel(c.onDataChannel)

	go func() {
		pingTicker := time.NewTicker(10 * time.Second)
		hudTicker := time.NewTicker(33 * time.Millisecond) //30hz
		sent := true
		hudToSend := models.Hud{}
		for {
			select {
			case <-c.Ctx.Done():
				log.Printf("stopping user updater: %s\n", c.Ctx.Err().Error())
				return
			case hud, ok := <-c.HudChannel:
				if !ok {
					log.Println("hud channel closed")
					return
				}
				if c.HudOutput != nil {
					hudToSend = hud
					sent = false
				}
			case <-pingTicker.C:
				if c.PingOutput != nil {
					data, err := json.Marshal(models.Ping{
						TimeStamp: time.Now().UnixMilli(),
						Source:    PingSourceName,
					})
					err = c.PingOutput.Send(data)
					if err != nil {
						log.Printf("error: failed sending ping: error - %s\n", err.Error())
						continue
					}
				}
			case <-hudTicker.C:
				if !sent && c.HudOutput != nil {
					encodedMsg, err := encode(hudToSend)
					sent = true
					err = c.HudOutput.SendText(encodedMsg)
					if err != nil {
						log.Printf("error: failed sending hud: error - %s\n", err.Error())
						continue
					}
				}
			}
		}
	}()
	return nil
}
