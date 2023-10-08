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

type VehicleStateIFace[T any] interface {
}

type VehicleSeat[T any] struct {
	lock sync.RWMutex
	seat *models.Seat

	seatCenterer      func(VehicleStateIFace[T]) VehicleStateIFace[T]
	seatCommandParser func(models.ControlState, models.ControlState, VehicleStateIFace[T]) VehicleStateIFace[T]
	hudUpdater        func(VehicleStateIFace[T]) models.Hud

	seatType string
	active   bool

	buttonMasks []uint32

	nextCommand     models.ControlState
	lastCommand     models.ControlState
	lastCommandTime time.Time
}

func NewVehicleSeat[T any](seat *models.Seat,
	parser func(models.ControlState, models.ControlState, VehicleStateIFace[T]) VehicleStateIFace[T],
	centerer func(VehicleStateIFace[T]) VehicleStateIFace[T],
	hudUpdater func(VehicleStateIFace[T]) models.Hud) *VehicleSeat[T] {
	return &VehicleSeat[T]{
		seat:              seat,
		seatCommandParser: parser,
		seatCenterer:      centerer,
		hudUpdater:        hudUpdater,
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

func (c *VehicleSeat[T]) UpdateHud(state VehicleStateIFace[T]) {
	if !c.active {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	select {
	case c.seat.HudChannel <- c.hudUpdater(state):
	default:
		log.Printf("%s seat hud channel full, skipping\n", c.seatType)
	}
}

func NewPress(oldState, newState models.ControlState, buttonIndex int, f func()) (bool, error) {
	if len(newState.Buttons) != len(oldState.Buttons) {
		return false, fmt.Errorf("length of buttons states mismatched")
	}

	if buttonIndex < 0 || buttonIndex > len(oldState.Buttons) {
		return false, fmt.Errorf("buttonIndex out of bounds - buttonIndex: %d maxIndex: %d", buttonIndex, len(oldState.Buttons))
	}

	if newState.Buttons[buttonIndex] && !oldState.Buttons[buttonIndex] {
		f()
		return true, nil
	}
	return false, nil
}

func GetValueWithMidDeadZone(value, midValue, deadZone float64) float64 {
	if value > midValue && midValue+deadZone > value {
		return midValue
	} else if value < midValue && midValue-deadZone < value {
		return midValue
	}
	return value
}

func GetValueWithLowDeadZone(value, lowValue, deadZone float64) float64 {
	if value > lowValue && lowValue+deadZone > value {
		return lowValue
	}
	return value
}

func MapToRange(value, min, max, minReturn, maxReturn float64) float64 {
	mappedValue := (maxReturn-minReturn)*(value-min)/(max-min) + minReturn

	if mappedValue > maxReturn {
		return maxReturn
	} else if mappedValue < minReturn {
		return minReturn
	} else {
		return mappedValue
	}
}

func ParseButtons(bitButton uint32, masks []uint32) []bool {
	returnvalue := make([]bool, 32)
	for i := range masks {
		returnvalue[i] = ((bitButton & masks[i]) != 0) //Check if bitbutton and mask both have bits in same place
	}
	return returnvalue
}
