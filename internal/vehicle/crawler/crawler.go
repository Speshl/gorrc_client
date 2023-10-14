package crawler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/vehicle"
	"golang.org/x/sync/errgroup"
)

func NewCrawler(cfg config.CrawlerConfig, commandDriver vehicle.CommandDriverIFace, seats []models.Seat) *Crawler {
	log.Printf("setting up crawler with %d seats\n", len(seats))

	crawlerState := NewCrawlerState(cfg)
	return &Crawler{
		cfg:           cfg,
		commandDriver: commandDriver,
		state:         crawlerState,
		seats:         NewCrawlerSeats(seats),
	}
}

func NewCrawlerState(cfg config.CrawlerConfig) CrawlerState {
	return CrawlerState{
		Gear:  0,
		Esc:   0.0,
		Steer: 0.0,
		Pan:   0.0,
		Tilt:  0.0,

		Trigger:    0.0,
		TurretPan:  0.0,
		TurretTilt: 0.0,

		PanSpeed:  cfg.PanSpeed,
		TiltSpeed: cfg.TiltSpeed,

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

func NewCrawlerSeats(seats []models.Seat) []*vehicle.VehicleSeat[CrawlerState] {
	crawlerSeats := make([]*vehicle.VehicleSeat[CrawlerState], 0, len(seats))
	for i := range seats {
		switch i {
		case 0:
			log.Println("setting up driver seat")
			crawlerSeats = append(crawlerSeats, NewDriverSeat(&seats[i]))
		case 1:
			log.Println("setting up passenger seat")
			crawlerSeats = append(crawlerSeats, NewPassengerSeat(&seats[i]))
		}
	}
	return crawlerSeats
}

func (c *Crawler) Init() error {
	err := c.commandDriver.Init()
	if err != nil {
		return fmt.Errorf("error: failed initializing crawler command interface: %w", err)
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

func (c *Crawler) Stop() error {
	log.Println("stopping crawler")
	err := c.commandDriver.Stop()
	if err != nil {
		return fmt.Errorf("error: failed stopping command driver: %w", err)
	}
	return nil
}

func (c *Crawler) Start(ctx context.Context) error {
	log.Println("starting crawler")
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
		for {
			select {
			case <-errGroupCtx.Done():
				log.Printf("stopping crawer state syncer: %s\n", ctx.Err().Error())
				return ctx.Err()
			case <-commandTicker.C:
				statesWithNewCommand := make([]CrawlerState, 0, len(c.seats))
				for i := range c.seats {
					newState := c.seats[i].ApplyCommand(c.state).(CrawlerState)
					statesWithNewCommand = append(statesWithNewCommand, newState)
				}

				mixedState := c.mergeSeatStates(statesWithNewCommand)
				err := c.applyState(mixedState)
				if err != nil {
					return fmt.Errorf("failed applying crawler state: %w", err)
				}

				for i := range c.seats {
					c.seats[i].UpdateHud(c.state)
				}
			}
		}
	})

	err := errGroup.Wait()
	if err != nil {
		return fmt.Errorf("crawler error group closed: %w", err)
	}
	return nil
}

// mergeSeatStates merges multiple states into one state. For cases where two seats have control over 1 axis, you can determine mixing here
func (c *Crawler) mergeSeatStates(states []CrawlerState) CrawlerState {
	if len(states) < 1 {
		log.Println("no crawler states given, so making an empty one")
		return NewCrawlerState(c.cfg)
	}

	return states[0] //TODO actually merge instead of just taking driver commands
}

func (c *Crawler) applyState(state CrawlerState) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.state = state

	commands := c.buildCommands(c.state)
	err := c.commandDriver.SetMany(commands)
	if err != nil {
		return fmt.Errorf("failed setting crawler commands: %w", err)
	}

	return nil
}

func (c *Crawler) buildCommands(state CrawlerState) []vehicle.DriverCommand {
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
		{
			Name:  "pan",
			Value: state.Pan,
			Min:   MinOutput,
			Max:   MaxOutput,
		},
		{
			Name:  "tilt",
			Value: state.Tilt,
			Min:   MinOutput,
			Max:   MaxOutput,
		},

		{
			Name:  "trigger",
			Value: state.Trigger,
			Min:   MinOutput,
			Max:   MaxOutput,
		},
		{
			Name:  "turret_pan",
			Value: state.TurretPan,
			Min:   MinOutput,
			Max:   MaxOutput,
		},
		{
			Name:  "turrent_tilt",
			Value: state.TurretTilt,
			Min:   MinOutput,
			Max:   MaxOutput,
		},
	}
}
