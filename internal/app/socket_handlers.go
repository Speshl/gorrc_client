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

	if offer.SeatNumber < a.cfg.ServerCfg.SeatCount || offer.SeatNumber > a.cfg.ServerCfg.SeatCount {
		log.Printf("offer was for unsupported seat number: %d\n", offer.SeatNumber)
		return
	}

	newConnection, err := NewConnection(context.Background(), socketConn, a.seats[offer.SeatNumber].CommandChannel, a.seats[offer.SeatNumber].HudChannel, a.speaker.TrackPlayer)
	if err != nil {
		log.Printf("failed creating connection on offer for seat %d: %s\n", offer.SeatNumber, err.Error())
		return
	}
	a.userConns[offer.SeatNumber] = newConnection

	err = a.userConns[offer.SeatNumber].RegisterHandlers(a.seats[offer.SeatNumber].AudioTracks, a.seats[offer.SeatNumber].VideoTracks)
	if err != nil {
		log.Printf("failed registering handelers for connection for seat %d: %s\n", offer.SeatNumber, err.Error())
		return
	}

	// Set the received offer as the remote description
	err = a.userConns[offer.SeatNumber].PeerConnection.SetRemoteDescription(offer.Offer)
	if err != nil {
		log.Printf("failed to set remote description: %s\n", err)
		return
	}

	// Create answer
	answer, err := a.userConns[offer.SeatNumber].PeerConnection.CreateAnswer(nil)
	if err != nil {
		log.Printf("Failed to create answer: %s\n", err)
		return
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(a.userConns[offer.SeatNumber].PeerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = a.userConns[offer.SeatNumber].PeerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println("Failed to set local description:", err)
		return
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	answerReq := models.Answer{
		Answer:     a.userConns[offer.SeatNumber].PeerConnection.LocalDescription(),
		SeatNumber: offer.SeatNumber,
	}

	encodedAnswer, err := encode(answerReq)
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
	log.Printf("car connected as %s(%s) @ %s(%s) with %d seats available\n", a.carInfo.Name, a.carInfo.ShortName, a.trackInfo.Name, a.trackInfo.ShortName, a.cfg.ServerCfg.SeatCount)
}
