package smallracer

import (
	"sync"
	"time"

	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/vehicle"
)

const (
	MaxSeats = 2

	//Button Maps
	TrimLeft     = 0
	TrimRight    = 1
	CamCenter    = 2
	TurretCenter = 2

	//Sequential
	UpShift         = 3
	DownShift       = 4
	SwitchTransType = 5

	//HPattern
	ReverseGear = 6
	FirstGear   = 7
	SecondGear  = 8
	ThirdGear   = 9
	FourthGear  = 10
	FifthGear   = 11
	SixthGear   = 12

	VolumeMute = 20
	VolumeUp   = 21
	VolumeDown = 22

	//TransTypes
	TransTypeSequential = "sequential"
	TransTypeHPattern   = "hpattern"

	TopGear                 = 6
	MaxTimeSinceLastCommand = 500 * time.Millisecond

	MaxTrimPerCycle = .01

	MaxVolumePerCycle = 10

	MaxVolume = 100
	MinVolume = 0

	DeadZone = 0.01

	MaxInput  = 1.0
	MinInput  = -1.0
	MaxOutput = 1.0
	MinOutput = -1.0
)

type Ratio struct {
	Name string
	Max  float64
	Min  float64
}

type SmallRacer struct {
	cfg           config.SmallRacerConfig
	lock          sync.RWMutex
	seats         []*vehicle.VehicleSeat[SmallRacerState]
	state         SmallRacerState
	commandDriver vehicle.CommandDriverIFace

	//Transmission
	// Ratios    map[int]Ratio
	// TransType string //use goenum

	//sound stuff
	// Volume int
}

type SmallRacerState struct {
	Esc       float64
	Steer     float64
	SteerTrim float64
	Gear      int
	TransType string

	//Config
	SteerSpeed float64

	Ratios map[int]Ratio
}
