package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Speshl/gorrc_client/internal/cam"
	"github.com/Speshl/gorrc_client/internal/command/pca9685"
	pipwm "github.com/Speshl/gorrc_client/internal/command/pi_pwm"
	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/gst"
	"github.com/Speshl/gorrc_client/internal/mic"
	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/speaker"
	"github.com/Speshl/gorrc_client/internal/vehicle"
	"github.com/Speshl/gorrc_client/internal/vehicle/crawler"
	smallracer "github.com/Speshl/gorrc_client/internal/vehicle/smallRacer"
	"github.com/google/uuid"
	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
	"golang.org/x/sync/errgroup"
)

const DriverSeatNum = 0
const PassengerSeatNum = 1

type App struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	cfg       config.Config

	vehicle vehicle.Vehicle

	vehicleInfo models.Car
	trackInfo   models.Track

	client *socketio.Client

	speakerChannel chan string
	speaker        *speaker.Speaker
	mic            *mic.Mic
	cams           []*cam.Cam
	command        vehicle.CommandDriverIFace

	seats         []models.Seat //number of available connections to this vehicle
	userConns     []*Connection
	userPeerConns map[uuid.UUID]*webrtc.PeerConnection
}

func NewApp(cfg config.Config, client *socketio.Client) *App {
	ctx, cancel := context.WithCancel(context.Background())

	speakerChannel := make(chan string, 100)

	if cfg.ServerCfg.SeatCount < 1 || cfg.ServerCfg.SeatCount > 2 {
		cfg.ServerCfg.SeatCount = config.DefaultSeatCount
	}

	seats := make([]models.Seat, 0, cfg.ServerCfg.SeatCount)
	for i := 0; i < cfg.ServerCfg.SeatCount; i++ {
		seats = append(seats, models.Seat{
			Index:          i,
			CommandChannel: make(chan models.ControlState, 100),
			HudChannel:     make(chan models.Hud, 100),
			VideoTracks:    make([]*webrtc.TrackLocalStaticSample, 0, len(cfg.CamCfgs)),
			AudioTracks:    make([]*webrtc.TrackLocalStaticSample, 0, 1),
		})
	}

	return &App{
		cfg:            cfg,
		client:         client,
		ctx:            ctx,
		ctxCancel:      cancel,
		speakerChannel: speakerChannel,
		vehicle:        newVehicle(cfg, seats),
		seats:          seats,
		speaker:        speaker.NewSpeaker(cfg.SpeakerCfg, speakerChannel),
		cams:           make([]*cam.Cam, 0, len(cfg.CamCfgs)),
		userConns:      make([]*Connection, cfg.ServerCfg.SeatCount),
		userPeerConns:  make(map[uuid.UUID]*webrtc.PeerConnection, 2),
	}
}

func (a *App) RegisterHandlers() error {
	log.Println("registering handlers")
	a.client.OnEvent("reply", func(s socketio.Conn, msg string) {
		log.Println("Receive Message /reply: ", "reply", msg)
	})

	a.client.OnEvent("offer", a.onOffer)

	a.client.OnEvent("candidate", a.onICECandidate)

	a.client.OnEvent("register_success", a.onRegisterSuccess)

	log.Println("attempting to connect to server...")
	err := a.client.Connect() //Client must have atleast 1 event handler to work
	if err != nil {
		return fmt.Errorf("error: failed connecting to server - %w", err)
	}
	log.Println("connected to server")
	return nil
}

func (a *App) Start() error {
	group, groupCtx := errgroup.WithContext(a.ctx)
	log.Println("starting...")

	for i, camCfg := range a.cfg.CamCfgs {
		if camCfg.Enabled { //start enabled cameras
			cam, err := cam.NewCam(camCfg)
			if err != nil {
				return fmt.Errorf("error: failed creating cam %d: %w\n", i, err)
			}
			a.cams = append(a.cams, cam)

			for i := range a.seats { //add all camera video tracks to each seat
				a.seats[i].VideoTracks = append(a.seats[i].VideoTracks, cam.VideoTrack)
			}
		}
	}

	mic, err := mic.NewMic(a.cfg.MicCfg)
	if err != nil {
		return fmt.Errorf("error: failed creating mic: %w\n", err)
	}
	a.mic = mic
	for i := range a.seats { //add audio tracks to each seat
		a.seats[i].AudioTracks = append(a.seats[i].AudioTracks, a.mic.AudioTrack)
	}

	defer func() {
		log.Println("stopping...")
		a.client.Close()
	}()

	//Start gstreamer loops
	group.Go(func() error {
		go func() {
			log.Println("starting gstreamer main send recieve loops")
			gst.StartMainSendLoop() //Start gstreamer main send loop from main thread
			log.Println("starting gstreamer main recieve loops")
			gst.StartMainRecieveLoop() //Start gstreamer main recieve loop from main thread
		}()
		return nil
	})

	group.Go(func() error {
		return a.speaker.Start(groupCtx)
	})

	//kill listener
	group.Go(func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-signalChannel:
			log.Printf("received signal: %s\n", sig)
			a.ctxCancel()
			return fmt.Errorf("received signal: %s\n", sig)
		case <-groupCtx.Done():
			log.Printf("closing signal goroutine\n")
			return groupCtx.Err()
		}
	})

	//Start Cameras
	for i, cam := range a.cams {
		group.Go(func() error {
			log.Printf("starting camera %d\n", i)
			return cam.Start(groupCtx)
		})
	}

	// Start Mic
	group.Go(func() error {
		log.Printf("starting mic")
		a.mic.Start()
		select {
		case <-groupCtx.Done():
			log.Printf("stopping mic")
			return nil
		}
	})

	//Start car
	group.Go(func() error {
		log.Printf("Starting car")
		err := a.vehicle.Init()
		if err != nil {
			return err
		}

		return a.vehicle.Start(groupCtx)
	})

	//Send connect and send healthchecks
	group.Go(func() error {
		encodedMsg, _ := encode(models.ConnectReq{
			Key:       a.cfg.ServerCfg.Key,
			Password:  a.cfg.ServerCfg.Password,
			SeatCount: a.cfg.ServerCfg.SeatCount,
		})
		a.client.Emit("car_connect", encodedMsg)

		healthTicker := time.NewTicker(30 * time.Second)

		for {
			select {
			case <-groupCtx.Done():
				log.Println("health checker stopped")
				return groupCtx.Err()
			case <-healthTicker.C:
				//log.Println("server healthcheck: healthy")
				a.client.Emit("car_healthy", "")
			}
		}
	})

	if !a.cfg.ServerCfg.SilentStart {
		err = a.speaker.Play(groupCtx, "startup")
		if err != nil {
			log.Printf("failed playing startup sound: %s\n", err.Error())
		}
	}

	defer func() {
		if !a.cfg.ServerCfg.SilentShutdown {
			err = a.speaker.Play(context.WithoutCancel(groupCtx), "shutdown")
			if err != nil {
				log.Printf("failed playing shutdown sound: %s\n", err.Error())
			}
		}
	}()

	err = group.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("context was cancelled")
			return nil
		} else {
			return fmt.Errorf("server stopping due to error - %w", err)
		}
	}

	log.Println("shutting down")
	return a.client.Close()
}

func newCommand(cfg config.CommandConfig) vehicle.CommandDriverIFace {
	switch cfg.CommandDriver {
	case "pca9685":
		log.Println("command driver: pca9685")
		return pca9685.NewCommand(cfg)
	case "pipwm":
		log.Println("command driver: pipwm")
		return pipwm.NewCommand(cfg)
	default:
		log.Println("warning: no servo driver selected, defaulting to pi pwm")
		return pipwm.NewCommand(cfg)
	}
}

func newVehicle(cfg config.Config, seats []models.Seat) vehicle.Vehicle {
	switch cfg.SmallRacerCfg.VehicleType {
	case "crawler":
		return crawler.NewCrawler(cfg.CrawlerCfg, newCommand(cfg.CommandCfg), seats)
	case "smallracer":
		fallthrough
	default:
		return smallracer.NewSmallRacer(cfg.SmallRacerCfg, newCommand(cfg.CommandCfg), seats)
	}
}
