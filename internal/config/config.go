package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	MaxSupportedServos = 16
	MaxSupportedCams   = 2
	AppEnvBase         = "GORRC_"

	DefaultServer         = "127.0.0.1:8181"
	DefaultCarKey         = "c0b839e9-0962-4494-9840-4b8751e15d90" //TODO Remove after testing
	DefaultCarType        = "crawler"
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
)

type Config struct {
	ServerCfg  ServerConfig
	CommandCfg CommandConfig
	CamCfgs    []CamConfig
	SpeakerCfg SpeakerConfig
	MicCfg     MicConfig

	CrawlerCfg CrawlerConfig
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
	CarType       string
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
	GearRMin float64
	GearRMax float64
	Gear1Min float64
	Gear1Max float64
	Gear2Min float64
	Gear2Max float64
	Gear3Min float64
	Gear3Max float64
	Gear4Min float64
	Gear4Max float64
	Gear5Min float64
	Gear5Max float64
	Gear6Min float64
	Gear6Max float64
}

func GetConfig() Config {
	cfg := Config{
		ServerCfg:  GetServerConfig(),
		CommandCfg: GetCommandConfig(),
		CamCfgs:    GetCamConfig(),
		SpeakerCfg: GetSpeakerConfig(),
		MicCfg:     GetMicConfig(),

		//Vehicle specific configs
		CrawlerCfg: GetCrawlerConfig(),
	}

	log.Printf("app Config: \n%+v\n", cfg)
	return cfg
}

func GetServerConfig() ServerConfig {
	return ServerConfig{
		Server:         GetStringEnv("SERVER", DefaultServer),
		Key:            GetStringEnv("CARKEY", DefaultCarKey),
		Password:       GetStringEnv("CARPASSWORD", DefaultPassword),
		SeatCount:      GetIntEnv("SEATCOUNT", DefaultSeatCount),
		SilentStart:    GetBoolEnv("SILENTSTART", DefaultSilentStart),
		SilentShutdown: GetBoolEnv("SILENTSHUTDOWN", DefaultSilentShutdown),
		SilentConnect:  GetBoolEnv("SILENTCONNECT", DefaultSilentConnect),
	}
}

func GetCommandConfig() CommandConfig {
	commandCfg := CommandConfig{
		CommandDriver: GetStringEnv("SERVODRIVER", DefaultCommandDriver),
		CarType:       GetStringEnv("CARTYPE", DefaultCarType),
		Address:       DefaultAddress, //  GetStringEnv("ADDRESS", DefaultAddress),
		I2CDevice:     GetStringEnv("I2CDEVICE", DefaultI2CDevice),
		ServoCfgs:     make([]ServoConfig, 0, MaxSupportedServos),
	}

	for i := 0; i < MaxSupportedServos; i++ {
		envPrefix := fmt.Sprintf("SERVO%d_", i)
		servoCfg := ServoConfig{
			Name:     GetStringEnv(envPrefix+"NAME", ""),
			Channel:  GetIntEnv(envPrefix+"CHANNEL", i),
			MaxPulse: float64(GetIntEnv(envPrefix+"MAXPULSE", DefaultMaxPulse)),
			MinPulse: float64(GetIntEnv(envPrefix+"MINPULSE", DefaultMinPulse)),
			Inverted: GetBoolEnv(envPrefix+"INVERTED", DefaultInverted),
			Offset:   GetIntEnv(envPrefix+"MIDOFFSET", DefaultOffset),
		}

		if servoCfg.Name != "" {
			log.Printf("found config for servo: %s\n", servoCfg.Name)
			commandCfg.ServoCfgs = append(commandCfg.ServoCfgs, servoCfg)
		}
	}
	return commandCfg
}

func GetCamConfig() []CamConfig {
	camCfgs := make([]CamConfig, 0, MaxSupportedCams)
	for i := 0; i < MaxSupportedCams; i++ {
		camPrefix := fmt.Sprintf("CAM%d_", i)
		camCfgs = append(camCfgs, CamConfig{
			Enabled:        GetBoolEnv(camPrefix+"ENABLED", DefaultCamEnable),
			Device:         GetStringEnv(camPrefix+"DEVICE", DefaultSpeakerDevice),
			Width:          GetStringEnv(camPrefix+"WIDTH", DefaultWidth),
			Height:         GetStringEnv(camPrefix+"HEIGHT", DefaultHeight),
			Fps:            GetStringEnv(camPrefix+"FPS", DefaultFPS),
			VerticalFlip:   GetBoolEnv(camPrefix+"VFLIP", DefaultVerticalFlip),
			HorizontalFlip: GetBoolEnv(camPrefix+"HFLIP", DefaultHorizontalFlip),
			Profile:        GetStringEnv(camPrefix+"PROFILE", DefaultProfile),
			Mode:           GetStringEnv(camPrefix+"MODE", DefaultMode),
		})
		if i == 0 {
			camCfgs[i].Enabled = true //force cam 0 to always be on
		}
	}
	return camCfgs
}

func GetSpeakerConfig() SpeakerConfig {
	return SpeakerConfig{
		Enabled: GetBoolEnv("SPEAKERENABLED", DefaultSpeakerEnabled),
		Device:  GetStringEnv("SPEAKERDEVICE", DefaultSpeakerDevice),
		Volume:  GetStringEnv("SPEAKERVOLUME", DefaultSpeakerVolume),
	}
}

func GetMicConfig() MicConfig {
	return MicConfig{
		Enabled: GetBoolEnv("MICENABLED", DefaultMicEnabled),
		Device:  GetStringEnv("MICDEVICE", DefaultMicDevice),
		Volume:  GetStringEnv("MICVOLUME", DefaultMicVolume),
	}
}

func GetCrawlerConfig() CrawlerConfig {
	envPrefix := "CRAWLER_"
	return CrawlerConfig{
		GearRMin: GetFloatEnv(envPrefix+"GEARR_MIN", DefaultCrawlerGearRMin),
		GearRMax: GetFloatEnv(envPrefix+"GEARR_MAX", DefaultCrawlerGearRMax),
		Gear1Min: GetFloatEnv(envPrefix+"GEAR1_MIN", DefaultCrawlerGear1Min),
		Gear1Max: GetFloatEnv(envPrefix+"GEAR1_MAX", DefaultCrawlerGear1Max),
		Gear2Min: GetFloatEnv(envPrefix+"GEAR2_MIN", DefaultCrawlerGear2Min),
		Gear2Max: GetFloatEnv(envPrefix+"GEAR2_MAX", DefaultCrawlerGear2Max),
		Gear3Min: GetFloatEnv(envPrefix+"GEAR3_MIN", DefaultCrawlerGear3Min),
		Gear3Max: GetFloatEnv(envPrefix+"GEAR3_MAX", DefaultCrawlerGear3Max),
		Gear4Min: GetFloatEnv(envPrefix+"GEAR4_MIN", DefaultCrawlerGear4Min),
		Gear4Max: GetFloatEnv(envPrefix+"GEAR4_MAX", DefaultCrawlerGear4Max),
		Gear5Min: GetFloatEnv(envPrefix+"GEAR5_MIN", DefaultCrawlerGear5Min),
		Gear5Max: GetFloatEnv(envPrefix+"GEAR5_MAX", DefaultCrawlerGear5Max),
		Gear6Min: GetFloatEnv(envPrefix+"GEAR6_MIN", DefaultCrawlerGear6Min),
		Gear6Max: GetFloatEnv(envPrefix+"GEAR6_MAX", DefaultCrawlerGear6Max),
	}
}

func GetIntEnv(env string, defaultValue int) int {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseInt(strings.Trim(envValue, "\r"), 10, 32)
		if err != nil {
			log.Printf("warning:%s not parsed - error: %s\n", env, err)
			return defaultValue
		} else {
			return int(value)
		}
	}
}

func GetBoolEnv(env string, defaultValue bool) bool {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseBool(strings.Trim(envValue, "\r"))
		if err != nil {
			log.Printf("warning:%s not parsed - error: %s\n", env, err)
			return defaultValue
		} else {
			return value
		}
	}
}

func GetStringEnv(env string, defaultValue string) string {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		return strings.ToLower(strings.Trim(envValue, "\r"))
	}
}

func GetFloatEnv(env string, defaultValue float64) float64 {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseFloat(envValue, 64)
		if err != nil {
			return defaultValue
		}
		return value
	}
}
