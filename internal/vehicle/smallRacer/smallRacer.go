package smallracer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/vehicle"
	"github.com/prometheus/procfs"
	"golang.org/x/sync/errgroup"
)

func NewSmallRacer(cfg config.SmallRacerConfig, commandDriver vehicle.CommandDriverIFace, seats []models.Seat) *SmallRacer {
	log.Printf("setting up small racer with %d seats\n", len(seats))

	state := NewSmallRacerState(cfg, TransTypeSequential)
	return &SmallRacer{
		cfg:           cfg,
		commandDriver: commandDriver,
		state:         state,
		seats:         NewSmallRacerSeats(seats),
	}
}

func NewSmallRacerState(cfg config.SmallRacerConfig, transType string) SmallRacerState {
	return SmallRacerState{
		Gear:       0,
		Esc:        0.0,
		Steer:      0.0,
		SteerSpeed: cfg.SteerSpeed,
		TransType:  transType,

		Ratios: map[int]Ratio{
			-1: {
				Name: "R",
				Max:  cfg.GearRMax,
				Min:  cfg.GearRMin,
			},
			0: {
				Name: "N",
				Max:  0.0,
				Min:  0.0,
			},
			1: {
				Name: "1",
				Max:  cfg.Gear1Max,
				Min:  cfg.Gear1Min,
			},
			2: {
				Name: "2",
				Max:  cfg.Gear2Max,
				Min:  cfg.Gear2Min,
			},
			3: {
				Name: "3",
				Max:  cfg.Gear3Max,
				Min:  cfg.Gear3Min,
			},
			4: {
				Name: "4",
				Max:  cfg.Gear4Max,
				Min:  cfg.Gear4Min,
			},
			5: {
				Name: "5",
				Max:  cfg.Gear5Max,
				Min:  cfg.Gear5Min,
			},
			6: {
				Name: "6",
				Max:  cfg.Gear6Max,
				Min:  cfg.Gear6Min,
			},
		},
	}
}

func NewSmallRacerSeats(seats []models.Seat) []*vehicle.VehicleSeat[SmallRacerState] {
	vehicleSeats := make([]*vehicle.VehicleSeat[SmallRacerState], 0, len(seats))
	for i := range seats {
		switch i {
		case 0:
			log.Println("setting up driver seat")
			vehicleSeats = append(vehicleSeats, NewDriverSeat(&seats[i]))
		case 1:
			log.Println("setting up passenger seat")
			vehicleSeats = append(vehicleSeats, NewPassengerSeat(&seats[i]))
		}
	}
	return vehicleSeats
}

func (c *SmallRacer) Init() error {
	err := c.commandDriver.Init()
	if err != nil {
		return fmt.Errorf("error: failed initializing small racer command interface: %w", err)
	}

	for i := range c.seats {
		err = c.seats[i].Init()
		if err != nil {
			return err
		}
	}

	//Center up servos
	return c.applyState(c.state)
}

func (c *SmallRacer) Stop() error {
	log.Println("stopping small racer")
	err := c.commandDriver.Stop()
	if err != nil {
		return fmt.Errorf("error: failed stopping command driver: %w", err)
	}
	return nil
}

func (c *SmallRacer) Start(ctx context.Context) error {
	log.Println("starting small racer")
	errGroup, errGroupCtx := errgroup.WithContext(ctx)

	defer c.Stop()

	for i := range c.seats {
		seatNum := i
		errGroup.Go(func() error {
			return c.seats[seatNum].Start(errGroupCtx)
		})
	}

	errGroup.Go(func() error {
		commandTicker := time.NewTicker(33 * time.Millisecond)
		p, err := procfs.Self()
		if err != nil {
			return fmt.Errorf("error: procfs could not get process: %w", err)
		}
		for {
			select {
			case <-errGroupCtx.Done():
				log.Printf("stopping small racer state syncer: %s\n", ctx.Err().Error())
				return ctx.Err()
			case <-commandTicker.C:
				statesWithNewCommand := make([]SmallRacerState, 0, len(c.seats))
				for i := range c.seats {
					newState := c.seats[i].ApplyCommand(c.state).(SmallRacerState)
					statesWithNewCommand = append(statesWithNewCommand, newState)
				}

				mixedState := c.mergeSeatStates(statesWithNewCommand)
				err := c.applyState(mixedState)
				if err != nil {
					return fmt.Errorf("failed applying small racer state: %w", err)
				}

				netDev, err := p.NetDev() //update network stats
				if err != nil {
					return fmt.Errorf("error: failed getting netstat: %w", err)
				}

				wlan0Stats, ok := netDev["wlan0"] //TODO make configurable
				if !ok {
					return fmt.Errorf("error: failed getting wlan0 stats: not found")
				}

				for i := range c.seats {
					c.seats[i].UpdateHud(c.state, wlan0Stats)
				}
			}
		}
	})

	err := errGroup.Wait()
	if err != nil {
		return fmt.Errorf("small racer error group closed: %w", err)
	}
	return nil
}

// mergeSeatStates merges multiple states into one state. For cases where two seats have control over 1 axis, you can determine mixing here
func (c *SmallRacer) mergeSeatStates(states []SmallRacerState) SmallRacerState {
	if len(states) < 1 {
		log.Println("no small racer states given, so making an empty one")
		return NewSmallRacerState(c.cfg, c.state.TransType)
	}

	return states[0] //TODO actually merge instead of just taking driver commands
}

func (c *SmallRacer) applyState(state SmallRacerState) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.state = state

	commands := c.buildCommands(c.state)
	err := c.commandDriver.SetMany(commands)
	if err != nil {
		return fmt.Errorf("failed setting small racer commands: %w", err)
	}

	return nil
}

func (c *SmallRacer) buildCommands(state SmallRacerState) []vehicle.DriverCommand {
	return []vehicle.DriverCommand{
		{
			Name:  "esc",
			Value: state.Esc,
			Min:   MinOutput,
			Max:   MaxOutput,
		},
		{
			Name:  "steer",
			Value: state.Steer,
			Min:   MinOutput,
			Max:   MaxOutput,
		},
	}
}
