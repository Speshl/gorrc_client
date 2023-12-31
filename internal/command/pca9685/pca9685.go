package pca9685

import (
	"fmt"
	"log"

	command "github.com/Speshl/gorrc_client/internal/command"
	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/vehicle"
	"github.com/googolgl/go-i2c"
	"github.com/googolgl/go-pca9685"
)

const (
	MaxValue = 1.0
	MinValue = 0.0
	MaxPulse = pca9685.ServoMaxPulseDef
	MinPulse = pca9685.ServoMinPulseDef
	AcRange  = pca9685.ServoRangeDef

	MaxSupportedServos = 16
)

type CommandDriver struct {
	cfg    config.CommandConfig
	servos map[string]Servo
	driver *pca9685.PCA9685
}

type Servo struct {
	name     string
	inverted bool
	offset   float64
	servo    *pca9685.Servo
}

func NewCommand(cfg config.CommandConfig) *CommandDriver {
	return &CommandDriver{
		cfg: cfg,
	}
}

func (c *CommandDriver) Stop() error {
	return nil
}

func (c *CommandDriver) Init() error {
	i2c, err := i2c.New(c.cfg.Address, c.cfg.I2CDevice)
	if err != nil {
		return fmt.Errorf("error starting i2c with address - %w", err)
	}

	c.driver, err = pca9685.New(i2c, nil)
	if err != nil {
		return fmt.Errorf("error getting servo driver - %w", err)
	}

	servos := make(map[string]Servo, MaxSupportedServos)
	for i := range c.cfg.ServoCfgs {
		name := c.cfg.ServoCfgs[i].Name
		servos[name] = Servo{
			name:     name,
			inverted: c.cfg.ServoCfgs[i].Inverted,
			offset:   float64(c.cfg.ServoCfgs[i].Offset) / 100,
			servo: c.driver.ServoNew(c.cfg.ServoCfgs[i].Channel, &pca9685.ServOptions{
				AcRange:  AcRange,
				MinPulse: float32(c.cfg.ServoCfgs[i].MinPulse),
				MaxPulse: float32(c.cfg.ServoCfgs[i].MaxPulse),
			}),
		}
		log.Printf("servo added: %s\n", name)
	}
	c.servos = servos
	c.CenterAll()
	return nil
}

func (c *CommandDriver) CenterAll() {
	log.Println("centering all servos")
	for i := range c.servos {
		c.servos[i].servo.Fraction(0.5)
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
		mappedValue := command.MapToRange(cmd.Value+val.offset, cmd.Min, cmd.Max, MinValue, MaxValue)
		if c.servos[cmd.Name].inverted {
			mappedValue = MaxValue - mappedValue
		}

		err := c.servos[cmd.Name].servo.Fraction(float32(mappedValue))
		if err != nil {
			return fmt.Errorf("failed setting servo value - name: %s value:  %.2f - error: %w\n", cmd.Name, mappedValue, err)
		}
	}
	return nil
}
