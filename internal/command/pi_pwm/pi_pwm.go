package pipwm

import (
	"fmt"
	"log"

	"github.com/Speshl/gorrc_client/internal/command"
	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/vehicle"
	"github.com/stianeikeland/go-rpio/v4"
)

const (
	Frequency          = 100000
	CycleLength        = uint32(2000)
	MaxSupportedServos = 2
)

var PinMap = []int{12, 13} //Servo0, Servo1

type CommandDriver struct {
	cfg    config.CommandConfig
	servos map[string]Servo
}

type Servo struct {
	name     string
	inverted bool
	offset   float64
	servo    rpio.Pin
	maxValue uint32
	minValue uint32
}

func NewCommand(cfg config.CommandConfig) *CommandDriver {
	return &CommandDriver{
		cfg: cfg,
	}
}

func (c *CommandDriver) Init() error {
	err := rpio.Open()
	if err != nil {
		return fmt.Errorf("failed opening rpio: %w", err)
	}

	servos := make(map[string]Servo, MaxSupportedServos)
	for i := range c.cfg.ServoCfgs {
		if i >= MaxSupportedServos {
			break
		}

		name := c.cfg.ServoCfgs[i].Name
		servos[name] = Servo{
			name:     name,
			inverted: c.cfg.ServoCfgs[i].Inverted,
			offset:   float64(c.cfg.ServoCfgs[i].Offset) / 100,
			servo:    rpio.Pin(PinMap[i]),
			maxValue: uint32(c.cfg.ServoCfgs[i].MaxPulse),
			minValue: uint32(c.cfg.ServoCfgs[i].MinPulse),
		}
		servos[name].servo.Mode(rpio.Pwm)
		servos[name].servo.Freq(Frequency)
		log.Printf("servo added: %s\n", name)
	}
	c.servos = servos
	c.CenterAll()
	return nil
}

func (c *CommandDriver) Stop() error {
	err := rpio.Close()
	if err != nil {
		return fmt.Errorf("failed closing rpio: %w", err)
	}
	return nil
}

func (c *CommandDriver) CenterAll() {
	log.Println("centering all servos")
	for i := range c.servos {
		midValue := (c.servos[i].maxValue + c.servos[i].minValue) / 2
		c.servos[i].servo.DutyCycle(midValue, CycleLength)
	}
}

func (c *CommandDriver) SetMany(cmds []vehicle.DriverCommand) error {
	for i := range cmds {
		err := c.Set(cmds[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CommandDriver) Set(cmd vehicle.DriverCommand) error {
	val, ok := c.servos[cmd.Name]
	if ok {
		mappedValue := command.MapToRange(cmd.Value+val.offset, cmd.Min, cmd.Max, float64(val.minValue), float64(val.maxValue))
		if c.servos[cmd.Name].inverted {
			mappedValue = float64(val.maxValue) - mappedValue
		}

		c.servos[cmd.Name].servo.DutyCycle(uint32(mappedValue), CycleLength)
	}
	return nil
}
