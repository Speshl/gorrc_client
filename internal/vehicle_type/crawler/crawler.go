package crawler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Speshl/gorrc_client/internal/command"
	"github.com/Speshl/gorrc_client/internal/models"
	vehicleType "github.com/Speshl/gorrc_client/internal/vehicle_type"
)

const (
	//Button Maps
	TrimLeft  = 0
	TrimRight = 1
	CamCenter = 2
	UpShift   = 3
	DownShift = 4

	VolumeMute = 20
	VolumeUp   = 21
	VolumeDown = 22

	//TransTypes
	TransTypeSequential = "sequential"
	TransTypeHPattern   = "hpattern"

	TopGear                 = 6
	MaxTimeSinceLastCommand = 500 * time.Millisecond

	MaxPanPerCycle  = 0.005
	MaxTiltPerCycle = 0.005

	MaxTrimPerCycle = .01

	MaxVolumePerCycle = 10

	MaxVolume = 100
	MinVolume = 0

	DeadZone = 0.05

	MaxInput  = 1.0
	MinInput  = -1.0
	MaxOutput = 1.0
	MinOutput = -1.0
)

var TransTypeMap = map[int]string{
	0: TransTypeSequential,
	1: TransTypeHPattern,
}

var GearRatios = map[int]Ratio{
	-1: {
		Name: "R",
		Max:  0.0,
		Min:  -0.4,
	},
	0: {
		Name: "N",
		Max:  0.0,
		Min:  0.0,
	},
	1: {
		Name: "1",
		Max:  0.1,
		Min:  -0.1,
	},
	2: {
		Name: "2",
		Max:  0.3,
		Min:  -0.2,
	},
	3: {
		Name: "3",
		Max:  0.5,
		Min:  -0.2,
	},
	4: {
		Name: "4",
		Max:  0.7,
		Min:  -0.2,
	},
	5: {
		Name: "5",
		Max:  0.9,
		Min:  -0.2,
	},
	6: {
		Name: "6",
		Max:  1.0,
		Min:  -0.2,
	},
}

type Ratio struct {
	Name string
	Max  float64
	Min  float64
}

type Crawler struct {
	command command.CommandIFace

	lock      sync.RWMutex
	Stopped   bool
	Esc       float64
	Steer     float64
	SteerTrim float64
	Pan       float64
	Tilt      float64

	ButtonMasks []uint32
	Buttons     []bool

	//Transmission
	Gear      int
	Ratios    map[int]Ratio
	TransType string //use goenum

	//sound stuff
	Volume int

	CommandChannel chan models.ControlState
	HudChannel     chan models.Hud

	LastCommand     *models.ControlState
	LastCommandTime time.Time
}

func NewCrawler(commandChan chan models.ControlState, hudChan chan models.Hud, command command.CommandIFace) *Crawler {
	return &Crawler{
		Stopped:        true,
		Gear:           0,
		Esc:            0.0,
		Steer:          0.0,
		Pan:            0.0,
		Tilt:           0.0,
		Volume:         50,
		ButtonMasks:    vehicleType.BuildButtonMasks(),
		Buttons:        make([]bool, 0, 32),
		Ratios:         GearRatios,
		HudChannel:     hudChan,
		CommandChannel: commandChan,

		command: command,
	}
}

func (c *Crawler) updateHud() {
	c.HudChannel <- models.Hud{
		Lines: []string{
			fmt.Sprintf("Steer:%.2f|Esc:%.2f|Pan:%.2f|Tilt:%.2f", c.Steer, c.Esc, c.Pan, c.Tilt),
			fmt.Sprintf("Trim:%.2f|Gear:%s|Vol:%d", c.SteerTrim, c.Ratios[c.Gear].Name, c.Volume),
		},
	}
}

func (c *Crawler) String() string {
	return fmt.Sprintf("Steer: %f | Esc: %f | Pan: %f | Tilt: %f | Trim: %f | Gear: %s | Volume: %d", c.Steer, c.Esc, c.Pan, c.Tilt, c.SteerTrim, c.Ratios[c.Gear].Name, c.Volume)
}

func (c *Crawler) Start(ctx context.Context) error {
	err := c.command.Init()
	if err != nil {
		return fmt.Errorf("failed initializing command interface: %w", err)
	}

	safetyTicker := time.NewTicker(MaxTimeSinceLastCommand)
	ctx, cancel := context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Printf("stopping safety monitor: %s\n", ctx.Err().Error())
			cancel()
			return ctx.Err()
		case <-safetyTicker.C:
			if time.Since(c.LastCommandTime) > MaxTimeSinceLastCommand {
				if !c.Stopped {
					log.Println("time since last command is to long, applying stop command")
					c.resetCar()
				}
			}
		case command, ok := <-c.CommandChannel:
			if !ok {
				log.Println("control state channel closed")
				cancel()
				return ctx.Err()
			}
			c.SetCommand(command)
		}
	}
}

func (c *Crawler) resetCar() {
	c.Steer = 0.0
	c.Esc = 0.0
	c.Pan = 0.0
	c.Tilt = 0.0

	c.SteerTrim = 0
	c.Gear = 0
	c.Volume = 50
	c.Stopped = true
}

func (c *Crawler) SetCommand(state models.ControlState) {
	c.lock.Lock()
	defer c.lock.Unlock()

	//Parse new state buttons
	state.Buttons = vehicleType.ParseButtons(state.BitButton, c.ButtonMasks)

	//if first time through just save state and wait for next
	if c.LastCommand == nil {
		c.LastCommand = &state
		return
	}
	c.Stopped = false

	//Handle buttons
	vehicleType.NewPress(*c.LastCommand, state, UpShift, c.upShift)
	vehicleType.NewPress(*c.LastCommand, state, DownShift, c.downShift)

	vehicleType.NewPress(*c.LastCommand, state, TrimLeft, c.trimLeft)
	vehicleType.NewPress(*c.LastCommand, state, TrimRight, c.trimRight)

	vehicleType.NewPress(*c.LastCommand, state, CamCenter, c.camCenter)

	vehicleType.NewPress(*c.LastCommand, state, VolumeMute, c.volumeMute)
	vehicleType.NewPress(*c.LastCommand, state, VolumeUp, c.volumeUp)
	vehicleType.NewPress(*c.LastCommand, state, VolumeDown, c.volumeDown)

	//Handle Axes
	c.mapSteer(state.Axes[0])
	c.mapEsc(state.Axes[1], state.Axes[2])
	c.mapPan(state.Axes[3])
	c.mapTilt(state.Axes[4])

	//log.Println(c.String())

	//Save the state to compare new state against next time
	c.LastCommand = &state
	c.LastCommandTime = time.Now()

	c.sendCommand()
	c.updateHud()
}

func (c *Crawler) sendCommand() {
	err := c.command.Set("steer", c.Steer, MinOutput, MaxOutput)
	if err != nil {
		log.Printf("failed setting %s servo value: %s", "steer", err.Error())
	}

	err = c.command.Set("esc", c.Esc, MinOutput, MaxOutput)
	if err != nil {
		log.Printf("failed setting %s servo value: %s", "steer", err.Error())
	}

	err = c.command.Set("pan", c.Pan, MinOutput, MaxOutput)
	if err != nil {
		log.Printf("failed setting %s servo value: %s", "steer", err.Error())
	}

	err = c.command.Set("tilt", c.Tilt, MinOutput, MaxOutput)
	if err != nil {
		log.Printf("failed setting %s servo value: %s", "steer", err.Error())
	}
}
