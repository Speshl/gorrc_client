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
	AppEnvBase         = "GORRC_"

	DefaultServer   = "127.0.0.1:8181"
	DefaultCarKey   = "c0b839e9-0962-4494-9840-4b8751e15d90" //TODO Remove after testing
	DefaultPassword = ""

	DefaultMaxPulse = 2250
	DefaultMinPulse = 750
	DefaultInverted = false
	DefaultOffset   = 0

	DefaultAddress   = 0x40
	DefaultI2CDevice = "/dev/i2c-1"
)

type Config struct {
	Server     string
	Key        string
	Password   string
	CommandCfg CommandConfig
}

type CommandConfig struct {
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

func GetConfig() Config {
	cfg := Config{
		Server:     GetStringEnv("SERVER", DefaultServer),
		Key:        GetStringEnv("CARKEY", DefaultCarKey),
		Password:   GetStringEnv("CARPASSWORD", DefaultPassword),
		CommandCfg: GetCommandConfig(),
	}

	log.Printf("app Config: \n%+v\n", cfg)
	return cfg
}

func GetCommandConfig() CommandConfig {
	commandCfg := CommandConfig{
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
			commandCfg.ServoCfgs = append(commandCfg.ServoCfgs, servoCfg)
		}
	}
	return commandCfg
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
		return strings.Trim(envValue, "\r")
	}
}
