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
	"github.com/Speshl/gorrc_client/internal/command"
	pca9685 "github.com/Speshl/gorrc_client/internal/command/pca9685"
	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/gst"
	"github.com/Speshl/gorrc_client/internal/models"
	vehicleType "github.com/Speshl/gorrc_client/internal/vehicle_type"
	"github.com/Speshl/gorrc_client/internal/vehicle_type/crawler"
	socketio "github.com/googollee/go-socket.io"
	"golang.org/x/sync/errgroup"
)

type App struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	car vehicleType.VehicleType

	carInfo   models.Car
	trackInfo models.Track

	client     *socketio.Client
	connection *Connection
	Cfg        config.Config

	commandChannel chan models.ControlState
	hudChannel     chan models.Hud

	// speaker *carspeaker.CarSpeaker
	// mic     *carmic.CarMic
	cam     *cam.Cam
	command command.CommandIFace
}

func NewApp(cfg config.Config, client *socketio.Client) *App {
	ctx, cancel := context.WithCancel(context.Background())

	commandChannel := make(chan models.ControlState, 100)
	hudChannel := make(chan models.Hud, 100)

	command := pca9685.NewCommand(cfg.CommandCfg)

	//TODO Support multiple car types
	return &App{
		Cfg:            cfg,
		client:         client,
		ctx:            ctx,
		ctxCancel:      cancel,
		commandChannel: commandChannel,
		hudChannel:     hudChannel,
		car:            crawler.NewCrawler(commandChannel, hudChannel, command),
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

	log.Println("attemping to connect to server...")
	err := a.client.Connect() //Client must have atleast 1 event handler to work
	if err != nil {
		return fmt.Errorf("error connecting to server - %w", err)
	}
	log.Println("connected to server")
	return nil
}

func (a *App) Start() error {
	group, groupCtx := errgroup.WithContext(a.ctx)
	log.Println("starting...")

	cam, err := cam.NewCam(a.Cfg.CamCfg)
	if err != nil {
		return fmt.Errorf("error creating carcam: %w\n", err)
	}
	a.cam = cam

	defer func() {
		log.Println("stopping...")
		a.client.Close()
	}()

	//Start gstreamer loops
	group.Go(func() error {
		log.Println("starting gstreamer main send recieve loops")
		gst.StartMainSendLoop() //Start gstreamer main send loop from main thread
		log.Println("starting gstreamer main recieve loops")
		gst.StartMainRecieveLoop() //Start gstreamer main recieve loop from main thread
		return fmt.Errorf("gstreamer pipelines stopped")
	})

	//kill listener
	group.Go(func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		select {
		case sig := <-signalChannel:
			fmt.Printf("received signal: %s\n", sig)
			a.ctxCancel()
		case <-groupCtx.Done():
			fmt.Printf("closing signal goroutine\n")
			return groupCtx.Err()
		}

		return nil
	})

	//Start Camera
	group.Go(func() error {
		return a.cam.Start(a.ctx)
	})

	//Start car
	group.Go(func() error {
		return a.car.Start(groupCtx)
	})

	//Send connect and send healthchecks
	group.Go(func() error {
		encodedMsg, _ := encode(models.ConnectReq{
			Key:      a.Cfg.Key,
			Password: a.Cfg.Password,
		})
		a.client.Emit("car_connect", encodedMsg)

		healthTicker := time.NewTicker(30 * time.Second)

		for {
			select {
			case <-groupCtx.Done():
				log.Println("health checker stopped")
				return groupCtx.Err()
			case <-healthTicker.C:
				log.Println("healthcheck: healthy")
				a.client.Emit("car_healthy", "")
			}
		}
	})

	err = group.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("context was cancelled")
			return nil
		} else {
			return fmt.Errorf("server stopping due to error - %w", err)
		}
	}

	return a.client.Close()
}
