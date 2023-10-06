package crawler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Speshl/gorrc_client/internal/models"
	vehicletype "github.com/Speshl/gorrc_client/internal/vehicle_type"
)

const saftyTime = 200 * time.Millisecond

func NewDriverSeat(seat *models.Seat) *CrawlerSeat {
	return &CrawlerSeat{
		seat:              seat,
		seatCommandParser: driverParser,
		seatType:          "driver",
		active:            false,
		buttonMasks:       vehicletype.BuildButtonMasks(),
	}
}

func NewPassengerSeat(seat *models.Seat) *CrawlerSeat {
	return &CrawlerSeat{
		seat:              seat,
		seatCommandParser: passengerParser,
		seatType:          "passenger",
		active:            false,
		buttonMasks:       vehicletype.BuildButtonMasks(),
	}
}

func (c *CrawlerSeat) Init() error {
	return nil
}

func (c *CrawlerSeat) Start(ctx context.Context) error {
	log.Printf("starting %s seat\n", c.seatType)

	saftyTicker := time.NewTicker(saftyTime)
	for {
		select {
		case <-ctx.Done():
			log.Printf("stopping %s seat state syncer: %s\n", c.seatType, ctx.Err().Error())
			return ctx.Err()
		case <-saftyTicker.C:
			c.lock.Lock()
			if c.active && time.Since(c.lastCommandTime) > saftyTime {
				//log.Printf("setting %s seat inactive due to time since last command\n", c.seatType)
				c.active = false
			}
			c.lock.Unlock()
		case command, ok := <-c.seat.CommandChannel:
			if !ok {
				return fmt.Errorf("%s seat command channel closed", c.seatType)
			}

			c.lock.Lock()
			if command.TimeStamp > c.nextCommand.TimeStamp {
				c.nextCommand = command
				c.lastCommandTime = time.Now()
				c.active = true
			}
			c.lock.Unlock()
		}
	}
}

func (c *CrawlerSeat) ApplyCommand(state CrawlerState) CrawlerState {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.active {
		c.nextCommand.Buttons = vehicletype.ParseButtons(c.nextCommand.BitButton, c.buttonMasks)
		if c.lastCommand.TimeStamp == 0 {
			log.Println("skipping first command")
			c.lastCommand = c.nextCommand
			return state
		}

		if c.nextCommand.TimeStamp-c.lastCommand.TimeStamp > 200 {
			log.Println("skipping command due to latency")
			c.lastCommand = c.nextCommand
			return state
		}

		newState := c.seatCommandParser(c.lastCommand, c.nextCommand, state)
		c.lastCommand = c.nextCommand
		return newState
	}
	return state
}

func driverParser(oldCommand, newCommand models.ControlState, crawlerState CrawlerState) CrawlerState {
	newState := crawlerState

	vehicletype.NewPress(oldCommand, newCommand, UpShift, newState.upShift)
	vehicletype.NewPress(oldCommand, newCommand, DownShift, newState.downShift)

	vehicletype.NewPress(oldCommand, newCommand, TrimLeft, newState.trimSteerLeft)
	vehicletype.NewPress(oldCommand, newCommand, TrimRight, newState.trimSteerRight)

	vehicletype.NewPress(oldCommand, newCommand, CamCenter, newState.camCenter)

	// vehicletype.NewPress(oldCommand, newCommand, VolumeMute, newState.volumeMute)
	// vehicletype.NewPress(oldCommand, newCommand, VolumeUp, newState.volumeUp)
	// vehicletype.NewPress(oldCommand, newCommand, VolumeDown, newState.volumeDown)

	newState.mapSteer(newCommand.Axes[0])
	newState.mapEsc(newCommand.Axes[1], newCommand.Axes[2])
	newState.mapPan(newCommand.Axes[3])
	newState.mapTilt(newCommand.Axes[4])

	return newState
}

func passengerParser(oldCommand, newCommand models.ControlState, crawlerState CrawlerState) CrawlerState {
	newState := crawlerState

	vehicletype.NewPress(oldCommand, newCommand, CamCenter, newState.turretCenter)

	newState.mapTurretPan(newCommand.Axes[3])
	newState.mapTurretTilt(newCommand.Axes[4])

	// vehicletype.NewPress(oldCommand, newCommand, VolumeMute, newState.volumeMute)
	// vehicletype.NewPress(oldCommand, newCommand, VolumeUp, newState.volumeUp)
	// vehicletype.NewPress(oldCommand, newCommand, VolumeDown, newState.volumeDown)

	return newState
}
