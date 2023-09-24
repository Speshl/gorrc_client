package command

import (
	"fmt"
	"log"

	"github.com/Speshl/gorrc_client/internal/config"
)

const (
	MaxValue = 1.0
	MinValue = -1.0

	MaxSupportedServos = 16
)

type CommandIFace interface {
	Set(string, float64, float64, float64) error
	Init() error
}

type TestCommand struct {
	cfg    config.CommandConfig
	servos map[string]TestServo
}

type TestServo struct {
	name     string
	inverted bool
	offset   float64
	value    float64
}

func NewTestCommand(cfg config.CommandConfig) *TestCommand {
	return &TestCommand{
		cfg: cfg,
	}
}

func (t *TestCommand) Init() error {
	t.servos = make(map[string]TestServo, len(t.servos))
	for i := range t.cfg.ServoCfgs {
		t.servos[t.cfg.ServoCfgs[i].Name] = TestServo{
			name:     t.cfg.ServoCfgs[i].Name,
			inverted: t.cfg.ServoCfgs[i].Inverted,
			offset:   float64(t.cfg.ServoCfgs[i].Offset) / 100,
		}
	}
	return nil
}

func (t *TestCommand) Set(name string, value, min, max float64) error {
	val, ok := t.servos[name]
	if ok {
		mappedValue := mapToRange(value+val.offset, min, max, MinValue, MaxValue)
		if t.servos[name].inverted {
			mappedValue = MaxValue - mappedValue
		}
		val.value = mappedValue
		t.servos[name] = val
		if name == "steer" {
			log.Printf("Servo %s: value: %.2f offset: %.2f mapped: %.2f\n", name, value, val.offset, mappedValue)
		}
		return nil
	}
	return fmt.Errorf("servo %s not found", name)
}

func mapToRange(value, min, max, minReturn, maxReturn float64) float64 {
	mappedValue := (maxReturn-minReturn)*(value-min)/(max-min) + minReturn

	if mappedValue > maxReturn {
		return maxReturn
	} else if mappedValue < minReturn {
		return minReturn
	} else {
		return mappedValue
	}
}
