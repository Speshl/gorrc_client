package vehicle

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Speshl/gorrc_client/internal/models"
)

const saftyTime = 200 * time.Millisecond

// type VehicleSeatIFace[T any] interface {
// 	Init() error
// 	Start(context.Context) error
// 	ApplyCommand(T) T
// }

type VehicleStateIFace[T any] interface {
}

type VehicleSeat[T any] struct {
	lock sync.RWMutex
	seat *models.Seat

	seatCenterer      func(VehicleStateIFace[T]) VehicleStateIFace[T]
	seatCommandParser func(models.ControlState, models.ControlState, VehicleStateIFace[T]) VehicleStateIFace[T]

	seatType string
	active   bool

	buttonMasks []uint32

	nextCommand     models.ControlState
	lastCommand     models.ControlState
	lastCommandTime time.Time
}

func NewVehicleSeat[T any](seat *models.Seat, parser func(models.ControlState, models.ControlState, VehicleStateIFace[T]) VehicleStateIFace[T], centerer func(VehicleStateIFace[T]) VehicleStateIFace[T]) *VehicleSeat[T] {
	return &VehicleSeat[T]{
		seat:              seat,
		seatCommandParser: parser,
		seatCenterer:      centerer,
		seatType:          "driver",
		active:            false,
		buttonMasks:       BuildButtonMasks(),
	}
}

func (c *VehicleSeat[T]) Init() error {
	return nil
}

func (c *VehicleSeat[T]) Start(ctx context.Context) error {
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
			if c.nextCommand.TimeStamp == 0 {
				c.nextCommand = command
			}

			if command.TimeStamp >= c.nextCommand.TimeStamp {
				c.nextCommand = command
				c.lastCommandTime = time.Now()
				c.active = true
			}
			c.lock.Unlock()
		}
	}
}

func (c *VehicleSeat[T]) ApplyCommand(state VehicleStateIFace[T]) VehicleStateIFace[T] {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.active {
		c.nextCommand.Buttons = ParseButtons(c.nextCommand.BitButton, c.buttonMasks)
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
	} else {
		return c.seatCenterer(state)
	}
}
