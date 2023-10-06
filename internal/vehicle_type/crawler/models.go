package crawler

import (
	"sync"
	"time"

	"github.com/Speshl/gorrc_client/internal/models"
	vehicletype "github.com/Speshl/gorrc_client/internal/vehicle_type"
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
	lock          sync.RWMutex
	seats         []vehicletype.VehicleSeatIFace[CrawlerState]
	state         CrawlerState
	commandDriver vehicletype.CommandDriverIFace

	buttonMasks []uint32
	buttons     []bool
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

type SeatCommandParser func(models.ControlState, models.ControlState, CrawlerState) CrawlerState

type CrawlerSeat struct {
	lock              sync.RWMutex
	seat              *models.Seat
	seatCommandParser SeatCommandParser
	seatType          string
	active            bool

	nextCommand     models.ControlState
	lastCommand     models.ControlState
	lastCommandTime time.Time
}
