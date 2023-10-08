package app

import (
	"log"

	"github.com/Speshl/gorrc_client/internal/models"
	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
)

func (a *App) onOffer(socketConn socketio.Conn, msgs []string) {
	log.Println("offer recieved")
	if len(msgs) != 1 {
		log.Printf("error: offer from %s had to many msgs: %d\n", socketConn.ID(), len(msgs))
		return
	}
	msg := msgs[0]

	offer := models.Offer{}
	err := decode(msg, &offer)
	if err != nil {
		log.Printf("error: offer from %s failed unmarshaling: %s\n - msg - %s", socketConn.ID(), err.Error(), string(msg))
		return
	}

	if offer.SeatNumber < 0 || offer.SeatNumber >= a.cfg.ServerCfg.SeatCount || offer.SeatNumber >= len(a.seats) {
		log.Printf("error: offer was for unsupported seat number: %d\n", offer.SeatNumber)
		return
	}

	newConnection, err := NewConnection(socketConn, a.seats[offer.SeatNumber].CommandChannel, a.seats[offer.SeatNumber].HudChannel, a.speaker.TrackPlayer)
	if err != nil {
		log.Printf("error: failed creating connection on offer for seat %d: %s\n", offer.SeatNumber, err.Error())
		return
	}
	a.userConns[offer.SeatNumber] = newConnection

	log.Println("registering handlers")

	err = a.userConns[offer.SeatNumber].RegisterHandlers(a.seats[offer.SeatNumber].AudioTracks, a.seats[offer.SeatNumber].VideoTracks)
	if err != nil {
		log.Printf("error: failed registering handelers for connection for seat %d: %s\n", offer.SeatNumber, err.Error())
		return
	}

	log.Println("setting remote description")

	// Set the received offer as the remote description
	err = a.userConns[offer.SeatNumber].PeerConnection.SetRemoteDescription(offer.Offer)
	if err != nil {
		log.Printf("error: failed to set remote description: %s\n", err)
		return
	}

	//log.Printf("remote description: %s\n", a.userConns[offer.SeatNumber].PeerConnection.RemoteDescription())

	log.Println("creating answer")

	// Create answer
	answer, err := a.userConns[offer.SeatNumber].PeerConnection.CreateAnswer(nil)
	if err != nil {
		log.Printf("error: failed to create answer: %s\n", err)
		return
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(a.userConns[offer.SeatNumber].PeerConnection)

	log.Println("setting local description")

	// Sets the LocalDescription, and starts our UDP listeners
	err = a.userConns[offer.SeatNumber].PeerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println("error: failed to set local description:", err)
		return
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	log.Println("waiting for ice gathering")
	<-gatherComplete
	log.Println("ice gathering complete")

	answerReq := models.Answer{
		Answer:     a.userConns[offer.SeatNumber].PeerConnection.LocalDescription(),
		SeatNumber: offer.SeatNumber,
	}

	encodedAnswer, err := encode(answerReq)
	if err != nil {
		log.Printf("error: failed encoding answer: %s", err.Error())
		return
	}
	log.Printf("accepting/answering offer for seat %d\n", offer.SeatNumber)
	a.client.Emit("answer", encodedAnswer)
}

func (a *App) onICECandidate(socketConn socketio.Conn, msg string) {
	decodedMsg := ""
	err := decode(msg, &decodedMsg)
	if err != nil {
		log.Printf("error: ice candidate from %s failed unmarshaling: %s\n", socketConn.ID(), string(msg))
		return
	}
}

func (a *App) onRegisterSuccess(socketConn socketio.Conn, msgs []string) {
	if len(msgs) != 1 {
		log.Printf("error: offer from %s had to many msgs: %d\n", socketConn.ID(), len(msgs))
	}
	msg := msgs[0]

	decodedMsg := models.ConnectResp{}
	err := decode(msg, &decodedMsg)
	if err != nil {
		log.Printf("error: ice candidate from %s failed unmarshaling: %s\n", socketConn.ID(), string(msg))
		return
	}

	a.carInfo = decodedMsg.Car
	a.trackInfo = decodedMsg.Track
	log.Printf("car connected as %s(%s) @ %s(%s) with %d seats available\n", a.carInfo.Name, a.carInfo.ShortName, a.trackInfo.Name, a.trackInfo.ShortName, a.cfg.ServerCfg.SeatCount)
}
