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
	PingInput  chan int64
}

func NewConnection(socketConn socketio.Conn, commandChan chan models.ControlState, hudChan chan models.Hud, speakers AudioPlayer, peerConn *webrtc.PeerConnection) (*Connection, error) {
	log.Printf("Creating User Connection %s\n", socketConn.ID())
	ctx, cancel := context.WithCancel(context.Background())
	conn := &Connection{
		// ID:             socketConn.ID(),
		Socket:         socketConn,
		PeerConnection: peerConn,
		Ctx:            ctx,
		CtxCancel:      cancel,
		CommandChannel: commandChan,
		HudChannel:     hudChan,
		Speaker:        speakers,
		PingInput:      make(chan int64, 10),
	}
	return conn, nil
}

func (c *Connection) Disconnect() {
	log.Println("user disconnecting")
	c.CtxCancel()
	c.PeerConnection.Close()
}

func (c *Connection) RegisterHandlers(audioTracks []*webrtc.TrackLocalStaticSample, videoTracks []*webrtc.TrackLocalStaticSample) error {

	log.Println("adding car audio tracks")
	_, err := c.PeerConnection.AddTrack(audioTracks[0]) //TODO add all audio tracks
	if err != nil {
		return fmt.Errorf("error adding audio track: %w", err)
	}

	log.Println("adding car video tracks")
	_, err = c.PeerConnection.AddTrack(videoTracks[0]) //TODO add all video tracks
	if err != nil {
		return fmt.Errorf("error adding video track: %w", err)
	}

	log.Println("set user audio track player")
	c.PeerConnection.OnTrack(c.Speaker) //TODO: Update this to kick out video tracks

	log.Println("start event listeners")
	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	c.PeerConnection.OnICEConnectionStateChange(c.onICEConnectionStateChange)

	// Handle ICE candidate messages from the client
	c.PeerConnection.OnICECandidate(c.onICECandidate)

	c.PeerConnection.OnDataChannel(c.onDataChannel)

	go func() { //TODO pull this out to somewhere else
		pingTicker := time.NewTicker(1 * time.Second)
		hudTicker := time.NewTicker(33 * time.Millisecond) //30hz
		sent := true
		hudToSend := models.Hud{}
		lastPing := int64(0)
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
			case recievedPing, ok := <-c.PingInput:
				if !ok {
					log.Println("ping channel closed")
					return
				}
				lastPing = recievedPing
			case <-hudTicker.C:
				if !sent && c.HudOutput != nil {
					if len(hudToSend.Lines) > 0 {
						hudToSend.Lines[0] = fmt.Sprintf("%s | Ping:%dms", hudToSend.Lines[0], lastPing)
					}
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
