package config

const (
	MaxSupportedServos = 16
	MaxSupportedCams   = 2
	AppEnvBase         = "GORRC_"

	DefaultServer         = "127.0.0.1:8181"
	DefaultCarKey         = "c0b839e9-0962-4494-9840-4b8751e15d90" //TODO Remove after testing
	DefaultVehicleType    = "smallracer"
	DefaultPassword       = ""
	DefaultSeatCount      = 1
	DefaultSilentStart    = false
	DefaultSilentConnect  = false
	DefaultSilentShutdown = false

	DefaultMaxPulse = 2250 //2000
	DefaultMinPulse = 750  //1000
	DefaultInverted = false
	DefaultOffset   = 0

	// Default Camera Options
	DefaultCamEnable      = false
	DefaultWidth          = "320"
	DefaultHeight         = "240"
	DefaultFPS            = "30"
	DefaultVerticalFlip   = false
	DefaultHorizontalFlip = false
	DefaultProfile        = "high"
	DefaultMode           = ""

	// Default Speaker Options
	DefaultSpeakerEnabled = false
	DefaultSpeakerDevice  = "0"
	DefaultSpeakerVolume  = "1.0"

	// Default Speaker Options
	DefaultMicEnabled = false
	DefaultMicDevice  = "1"
	DefaultMicVolume  = "1.0"

	// Default Command Options
	DefaultCommandDriver = "pca9685"
	DefaultAddress       = 0x40
	DefaultI2CDevice     = "/dev/i2c-1"

	//Vehicle Specific Configs
	DefaultCrawlerGearRMin = -0.40
	DefaultCrawlerGearRMax = 0.00
	DefaultCrawlerGear1Min = -0.10
	DefaultCrawlerGear1Max = 0.10
	DefaultCrawlerGear2Min = -0.20
	DefaultCrawlerGear2Max = 0.20
	DefaultCrawlerGear3Min = -0.40
	DefaultCrawlerGear3Max = 0.40
	DefaultCrawlerGear4Min = -0.60
	DefaultCrawlerGear4Max = 0.60
	DefaultCrawlerGear5Min = -0.80
	DefaultCrawlerGear5Max = 0.80
	DefaultCrawlerGear6Min = -1.00
	DefaultCrawlerGear6Max = 1.00

	DefaultCrawlerPanSpeed  = 1
	DefaultCrawlerTiltSpeed = 1

	DefaultSmallRacerSteerSpeed = 1.0 //linear
)

type Config struct {
	ServerCfg  ServerConfig
	CommandCfg CommandConfig
	CamCfgs    []CamConfig
	SpeakerCfg SpeakerConfig
	MicCfg     MicConfig

	CrawlerCfg    CrawlerConfig
	SmallRacerCfg SmallRacerConfig
}

type ServerConfig struct {
	Server         string
	Key            string
	Password       string
	SeatCount      int
	SilentStart    bool
	SilentShutdown bool
	SilentConnect  bool
}

type CommandConfig struct {
	CommandDriver string
	Address       byte
	I2CDevice     string
	ServoCfgs     []ServoConfig
}

type ServoConfig struct {
	Name     string
	Inverted bool
	Type     string
	Channel  int
	MaxPulse float64
	MinPulse float64
	DeadZone int
	Offset   int
}

type CamConfig struct {
	Enabled        bool
	Device         string
	Width          string
	Height         string
	Fps            string
	DisableVideo   bool
	HorizontalFlip bool
	VerticalFlip   bool
	DeNoise        bool
	Rotation       int
	Level          string
	Profile        string
	Mode           string
}

type SpeakerConfig struct {
	Enabled bool
	Device  string
	Volume  string
}

type MicConfig struct {
	Enabled bool
	Device  string
	Volume  string
}

type CrawlerConfig struct {
	VehicleConfig
	PanSpeed  float64
	TiltSpeed float64
}

type SmallRacerConfig struct {
	VehicleConfig
	SteerSpeed float64
}

type VehicleConfig struct {
	VehicleType string
	GearRMin    float64
	GearRMax    float64
	Gear1Min    float64
	Gear1Max    float64
	Gear2Min    float64
	Gear2Max    float64
	Gear3Min    float64
	Gear3Max    float64
	Gear4Min    float64
	Gear4Max    float64
	Gear5Min    float64
	Gear5Max    float64
	Gear6Min    float64
	Gear6Max    float64
}
