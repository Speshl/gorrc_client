package crawler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Speshl/gorrc_client/internal/models"
	vehicletype "github.com/Speshl/gorrc_client/internal/vehicle_type"
	"golang.org/x/sync/errgroup"
)

func NewCrawler(commandDriver vehicletype.CommandDriverIFace, seats []models.Seat) *Crawler {
	log.Printf("setting up crawler with %d seats\n", len(seats))

	crawlerState := NewCrawlerState()
	return &Crawler{
		commandDriver: commandDriver,
		state:         crawlerState,
		seats:         NewCrawlerSeats(seats),
	}
}

func NewCrawlerState() CrawlerState {
	return CrawlerState{
		Gear:  0,
		Esc:   0.0,
		Steer: 0.0,
		Pan:   0.0,
		Tilt:  0.0,

		Trigger:    0.0,
		TurretPan:  0.0,
		TurretTilt: 0.0,

		Ratios: GearRatios,
	}
}

func NewCrawlerSeats(seats []models.Seat) []vehicletype.VehicleSeatIFace[CrawlerState] {
	crawlerSeats := make([]vehicletype.VehicleSeatIFace[CrawlerState], 0, len(seats))
	for i := range seats {
		switch i {
		case 0:
			log.Println("setting up driver seat")
			crawlerSeats = append(crawlerSeats, NewDriverSeat(&seats[i]))
		case 1:
			log.Println("setting up passenger seat")
			//crawlerSeats = append(crawlerSeats, NewPassengerSeat(&seats[i]))
		}
	}
	return crawlerSeats
}

func (c *Crawler) Init() error {
	err := c.commandDriver.Init()
	if err != nil {
		return fmt.Errorf("failed initializing crawler command interface: %w", err)
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

func (c *Crawler) Start(ctx context.Context) error {
	log.Println("starting crawler")
	errGroup, errGroupCtx := errgroup.WithContext(ctx)

	for i := range c.seats {
		errGroup.Go(func() error {
			return c.seats[i].Start(errGroupCtx)
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
					newState := c.seats[i].ApplyCommand(c.state)
					statesWithNewCommand = append(statesWithNewCommand, newState)
				}

				mixedState := c.mergeSeatStates(statesWithNewCommand)
				err := c.applyState(mixedState)
				if err != nil {
					return fmt.Errorf("failed applying crawler state: %w", err)
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
		return NewCrawlerState()
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

func (c *Crawler) buildCommands(state CrawlerState) []vehicletype.DriverCommand {
	return []vehicletype.DriverCommand{
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
