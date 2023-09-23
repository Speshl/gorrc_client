package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const AppEnvBase = "GORRC_"

const DefaultServer = "127.0.0.1:8181"
const DefaultCarKey = "c0b839e9-0962-4494-9840-4b8751e15d90" //TODO Remove after testing
const DefaultPassword = ""

type Config struct {
	Server    string
	Key       string
	Password  string
	ServoCfgs []ServoCfg
}

type ServoCfg struct {
	Invert   bool
	Type     string
	Channel  int
	MaxPulse int
	MinPulse int
	DeadZone int
}

func GetConfig() Config {
	cfg := Config{
		Server:   GetStringEnv("SERVER", DefaultServer),
		Key:      GetStringEnv("CARKEY", DefaultCarKey),
		Password: GetStringEnv("CARPASSWORD", DefaultPassword),
	}

	log.Printf("app Config: \n%+v\n", cfg)
	return cfg
}

func GetIntEnv(env string, defaultValue int) int {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseInt(envValue, 10, 32)
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
		value, err := strconv.ParseBool(envValue)
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
