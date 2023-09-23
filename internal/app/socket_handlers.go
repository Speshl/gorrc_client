package app

import (
	"context"
	"log"

	"github.com/Speshl/gorrc_client/internal/models"
	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
)

func (a *App) onOffer(socketConn socketio.Conn, msgs []string) {
	if len(msgs) != 1 {
		log.Printf("offer from %s had to many msgs: %d\n", socketConn.ID(), len(msgs))
	}
	msg := msgs[0]

	offer := models.Offer{}
	err := decode(msg, &offer)
	if err != nil {
		log.Printf("offer from %s failed unmarshaling: %s\n - msg - %s", socketConn.ID(), err.Error(), string(msg))
		return
	}

	newConnection, err := NewConnection(context.Background(), socketConn, a.commandChannel, a.hudChannel)
	if err != nil {
		log.Printf("failed creating connection on offer: %s\n", err.Error())
		return
	}
	a.connection = newConnection

	err = a.connection.RegisterHandlers(nil, nil)
	if err != nil {
		log.Printf("failed registering handelers for connection")
		return
	}

	log.Printf("received offer size: %d\n", len(offer.Offer.SDP))

	// Set the received offer as the remote description
	err = a.connection.PeerConnection.SetRemoteDescription(offer.Offer)
	if err != nil {
		log.Printf("failed to set remote description: %s\n", err)
		return
	}

	// Create answer
	answer, err := a.connection.PeerConnection.CreateAnswer(nil)
	if err != nil {
		log.Printf("Failed to create answer: %s\n", err)
		return
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(a.connection.PeerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = a.connection.PeerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println("Failed to set local description:", err)
		return
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	encodedAnswer, err := encode(a.connection.PeerConnection.LocalDescription())
	if err != nil {
		log.Printf("Failed encoding answer: %s", err.Error())
		return
	}
	log.Println("sending answer")
	a.client.Emit("answer", encodedAnswer)
}

func (a *App) onICECandidate(socketConn socketio.Conn, msg string) {
	decodedMsg := ""
	err := decode(msg, &decodedMsg)
	if err != nil {
		log.Printf("ice candidate from %s failed unmarshaling: %s\n", socketConn.ID(), string(msg))
		return
	}
}

func (a *App) onRegisterSuccess(socketConn socketio.Conn, msgs []string) {
	if len(msgs) != 1 {
		log.Printf("offer from %s had to many msgs: %d\n", socketConn.ID(), len(msgs))
	}
	msg := msgs[0]

	decodedMsg := models.ConnectResp{}
	err := decode(msg, &decodedMsg)
	if err != nil {
		log.Printf("ice candidate from %s failed unmarshaling: %s\n", socketConn.ID(), string(msg))
		return
	}

	a.carInfo = decodedMsg.Car
	a.trackInfo = decodedMsg.Track
	log.Printf("car connected as %s(%s) @ %s(%s)\n", a.carInfo.Name, a.carInfo.ShortName, a.trackInfo.Name, a.trackInfo.ShortName)
}
