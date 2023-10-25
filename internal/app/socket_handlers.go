package app

import (
	"fmt"
	"log"

	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/google/uuid"
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

	peerConn, err := a.getOrCreatePeerConn(offer.UserId)
	if err != nil {
		log.Printf("error: failed getting or creating peer conn: %s\n", err.Error())
		return
	}

	newConnection, err := NewConnection(offer.SeatNumber, socketConn, a.seats[offer.SeatNumber].CommandChannel, a.seats[offer.SeatNumber].HudChannel, a.speaker.TrackPlayer, peerConn)
	if err != nil {
		log.Printf("error: failed creating connection on offer for seat %d: %s\n", offer.SeatNumber, err.Error())
		return
	}
	a.userConns[offer.SeatNumber] = newConnection

	log.Printf("registering handlers for seat %d\n", offer.SeatNumber)

	err = a.userConns[offer.SeatNumber].RegisterHandlers(a.seats[offer.SeatNumber].AudioTracks, a.seats[offer.SeatNumber].VideoTracks)
	if err != nil {
		log.Printf("error: failed registering handelers for seat %d: %s\n", offer.SeatNumber, err.Error())
		return
	}

	log.Println("setting remote description for seat %d\n", offer.SeatNumber)
	err = a.userConns[offer.SeatNumber].PeerConnection.SetRemoteDescription(offer.Offer)
	if err != nil {
		log.Printf("error: failed to set remote description for seat %d: %s\n", offer.SeatNumber, err)
		return
	}

	log.Println("creating answer for seat %d\n", offer.SeatNumber)
	answer, err := a.userConns[offer.SeatNumber].PeerConnection.CreateAnswer(nil)
	if err != nil {
		log.Printf("error: failed to create answer for seat %d: %s\n", offer.SeatNumber, err)
		return
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(a.userConns[offer.SeatNumber].PeerConnection)

	log.Printf("setting local description for seat %d\n", offer.SeatNumber)

	// Sets the LocalDescription, and starts our UDP listeners
	err = a.userConns[offer.SeatNumber].PeerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Printf("error: failed to set local description for seat %d: %s\n", offer.SeatNumber, err.Error())
		return
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	log.Printf("waiting for ice gathering for seat %d\n", offer.SeatNumber)
	<-gatherComplete
	log.Printf("ice gathering complete for seat %d\n", offer.SeatNumber)

	answerReq := models.Answer{
		Answer:     a.userConns[offer.SeatNumber].PeerConnection.LocalDescription(),
		SeatNumber: offer.SeatNumber,
	}

	encodedAnswer, err := encode(answerReq)
	if err != nil {
		log.Printf("error: failed encoding answer for seat %d: %s", err.Error(), offer.SeatNumber)
		return
	}
	log.Printf("accepting/answering offer for seat %d\n", offer.SeatNumber)
	a.client.Emit("answer", encodedAnswer)
}

func (a *App) onICECandidate(socketConn socketio.Conn, msgs []string) {
	log.Println("ice candidate recieved")
	if len(msgs) != 1 {
		log.Printf("error: ice candidate from %s had to many msgs: %d\n", socketConn.ID(), len(msgs))
		return
	}
	msg := msgs[0]

	var userIceCandidate models.IceCandidate
	err := decode(msg, &userIceCandidate)
	if err != nil {
		log.Printf("error: ice candidate from %s failed unmarshaling: %s\n", socketConn.ID(), string(msg))
		return
	}

	peerConn, err := a.getOrCreatePeerConn(userIceCandidate.UserId)
	if err != nil {
		log.Printf("error: failed getting or creating peer conn: %s\n", err.Error())
		return
	}

	if userIceCandidate.Candidate.Candidate == "" {
		log.Println("warning: recieved empty ice candidate")
		return
	}

	err = peerConn.AddICECandidate(userIceCandidate.Candidate)
	if err != nil {
		log.Printf("error: failed to add ice candidate for user %s: %s\n", userIceCandidate.UserId.String(), err.Error())
		return
	}

	//a.userPeerConns[userIceCandidate.UserId] = peerConn

	log.Printf("recieved ice candidate: %s\n", userIceCandidate.Candidate.Candidate)
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

	a.vehicleInfo = decodedMsg.Car
	a.trackInfo = decodedMsg.Track
	log.Printf("car connected as %s(%s) @ %s(%s) with %d seats available\n", a.vehicleInfo.Name, a.vehicleInfo.ShortName, a.trackInfo.Name, a.trackInfo.ShortName, a.cfg.ServerCfg.SeatCount)
}

func (a *App) getOrCreatePeerConn(userId uuid.UUID) (*webrtc.PeerConnection, error) {
	var err error

	peerConn, ok := a.userPeerConns[userId]
	if !ok || peerConn.ConnectionState() == webrtc.PeerConnectionStateClosed {
		log.Printf("creating new peer connection for user %s\n", userId)
		peerConn, err = webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error: failed creating peer connection for user %s on ice candidate: %w\n", userId, err)
		}
		a.userPeerConns[userId] = peerConn
	}
	return peerConn, nil
}
