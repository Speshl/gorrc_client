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

	DefaultServer    = "127.0.0.1:8181"
	DefaultCarKey    = "c0b839e9-0962-4494-9840-4b8751e15d90" //TODO Remove after testing
	DefaultCarType   = "crawler"
	DefaultPassword  = ""
	DefaultSeatCount = 1

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
	DefaultSpeakerDevice = "0"
	DefaultSpeakerVolume = "1.0"

	// Default Speaker Options
	DefaultMicDevice = "1"
	DefaultMicVolume = "1.0"

	// Default Command Options
	DefaultAddress   = 0x40
	DefaultI2CDevice = "/dev/i2c-1"
)

type Config struct {
	ServerCfg  ServerConfig
	CommandCfg CommandConfig
	CamCfgs    []CamConfig
	SpeakerCfg SpeakerConfig
	MicCfg     MicConfig
}

type ServerConfig struct {
	Server    string
	Key       string
	Password  string
	SeatCount int
}

type CommandConfig struct {
	CarType   string
	Address   byte
	I2CDevice string
	ServoCfgs []ServoConfig
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
	Device string
	Volume string
}

type MicConfig struct {
	Device string
	Volume string
}

func GetConfig() Config {
	cfg := Config{
		ServerCfg:  GetServerConfig(),
		CommandCfg: GetCommandConfig(),
		CamCfgs:    GetCamConfig(),
		SpeakerCfg: GetSpeakerConfig(),
		MicCfg:     GetMicConfig(),
	}

	log.Printf("app Config: \n%+v\n", cfg)
	return cfg
}

func GetServerConfig() ServerConfig {
	return ServerConfig{
		Server:    GetStringEnv("SERVER", DefaultServer),
		Key:       GetStringEnv("CARKEY", DefaultCarKey),
		Password:  GetStringEnv("CARPASSWORD", DefaultPassword),
		SeatCount: GetIntEnv("SEATCOUNT", DefaultSeatCount),
	}
}

func GetCommandConfig() CommandConfig {
	commandCfg := CommandConfig{
		CarType:   GetStringEnv("CARTYPE", DefaultCarType),
		Address:   DefaultAddress, //  GetStringEnv("ADDRESS", DefaultAddress),
		I2CDevice: GetStringEnv("I2CDEVICE", DefaultI2CDevice),
		ServoCfgs: make([]ServoConfig, 0, MaxSupportedServos),
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
		Device: GetStringEnv("SPEAKERDEVICE", DefaultSpeakerDevice),
		Volume: GetStringEnv("SPEAKERVOLUME", DefaultSpeakerVolume),
	}
}

func GetMicConfig() MicConfig {
	return MicConfig{
		Device: GetStringEnv("MICDEVICE", DefaultMicDevice),
		Volume: GetStringEnv("MICVOLUME", DefaultMicVolume),
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
