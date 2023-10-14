package crawler

import (
	"sync"
	"time"

	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/vehicle"
)

const (
	//Button Maps
	MaxSeats     = 2
	TrimLeft     = 0
	TrimRight    = 1
	CamCenter    = 2
	TurretCenter = 2
	UpShift      = 3
	DownShift    = 4

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

type Ratio struct {
	Name string
	Max  float64
	Min  float64
}

type Crawler struct {
	cfg           config.CrawlerConfig
	lock          sync.RWMutex
	seats         []*vehicle.VehicleSeat[CrawlerState]
	state         CrawlerState
	commandDriver vehicle.CommandDriverIFace

	//Transmission
	// Ratios    map[int]Ratio
	// TransType string //use goenum

	//sound stuff
	// Volume int
}

type CrawlerState struct {
	Esc       float64
	Steer     float64
	SteerTrim float64
	Pan       float64
	Tilt      float64
	Gear      int

	Trigger    float64
	TurretPan  float64
	TurretTilt float64

	Ratios map[int]Ratio
}
