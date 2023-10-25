package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func GetConfig() Config {
	cfg := Config{
		ServerCfg:  GetServerConfig(),
		CommandCfg: GetCommandConfig(),
		CamCfgs:    GetCamConfig(),
		SpeakerCfg: GetSpeakerConfig(),
		MicCfg:     GetMicConfig(),

		//Vehicle specific configs
		CrawlerCfg:    GetCrawlerConfig(),
		SmallRacerCfg: GetSmallRacerConfig(),
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
		PanSpeed:      GetFloatEnv(envPrefix+"PAN_SPEED", DefaultCrawlerPanSpeed),
		TiltSpeed:     GetFloatEnv(envPrefix+"TILT_SPEED", DefaultCrawlerTiltSpeed),
		VehicleConfig: GetVehicleConfig(),
	}
}

func GetSmallRacerConfig() SmallRacerConfig {
	envPrefix := "SMALLRACER_"
	return SmallRacerConfig{
		SteerSpeed:    GetFloatEnv(envPrefix+"STEER_SPEED", DefaultSmallRacerSteerSpeed),
		VehicleConfig: GetVehicleConfig(),
	}
}

func GetVehicleConfig() VehicleConfig {
	return VehicleConfig{
		VehicleType: GetStringEnv("VEHICLETYPE", DefaultVehicleType),
		GearRMin:    GetFloatEnv("GEARR_MIN", DefaultCrawlerGearRMin),
		GearRMax:    GetFloatEnv("GEARR_MAX", DefaultCrawlerGearRMax),
		Gear1Min:    GetFloatEnv("GEAR1_MIN", DefaultCrawlerGear1Min),
		Gear1Max:    GetFloatEnv("GEAR1_MAX", DefaultCrawlerGear1Max),
		Gear2Min:    GetFloatEnv("GEAR2_MIN", DefaultCrawlerGear2Min),
		Gear2Max:    GetFloatEnv("GEAR2_MAX", DefaultCrawlerGear2Max),
		Gear3Min:    GetFloatEnv("GEAR3_MIN", DefaultCrawlerGear3Min),
		Gear3Max:    GetFloatEnv("GEAR3_MAX", DefaultCrawlerGear3Max),
		Gear4Min:    GetFloatEnv("GEAR4_MIN", DefaultCrawlerGear4Min),
		Gear4Max:    GetFloatEnv("GEAR4_MAX", DefaultCrawlerGear4Max),
		Gear5Min:    GetFloatEnv("GEAR5_MIN", DefaultCrawlerGear5Min),
		Gear5Max:    GetFloatEnv("GEAR5_MAX", DefaultCrawlerGear5Max),
		Gear6Min:    GetFloatEnv("GEAR6_MIN", DefaultCrawlerGear6Min),
		Gear6Max:    GetFloatEnv("GEAR6_MAX", DefaultCrawlerGear6Max),
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
