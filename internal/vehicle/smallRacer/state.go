package smallracer

import (
	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/vehicle"
)

func (c *SmallRacerState) upShift() {
	if c.Gear < TopGear {
		c.Gear++
	}
}

func (c *SmallRacerState) downShift() {
	if c.Gear > -1 {
		c.Gear--
	}
}

func (c *SmallRacerState) switchTransType() {
	switch c.TransType {
	case TransTypeHPattern:
		c.TransType = TransTypeSequential
	case TransTypeSequential:
		c.TransType = TransTypeHPattern
	}
}

func (c *SmallRacerState) trimSteerLeft() {
	if c.SteerTrim-MaxTrimPerCycle < MinInput {
		c.SteerTrim = MinInput
	} else {
		c.SteerTrim -= MaxTrimPerCycle
	}
}

func (c *SmallRacerState) trimSteerRight() {
	if c.SteerTrim+MaxTrimPerCycle > MaxInput {
		c.SteerTrim = MaxInput
	} else {
		c.SteerTrim += MaxTrimPerCycle
	}
}

func (c *SmallRacerState) mapHPattern(newState models.ControlState, forwardGears []int, reverseGear int) {
	if c.TransType != TransTypeHPattern {
		return
	}

	//check if in a forward gear
	for i, forwardGear := range forwardGears {
		if newState.Buttons[forwardGear] {
			if i+1 <= TopGear {
				c.Gear = i + 1
			} else {
				c.Gear = TopGear
			}
			return
		}
	}

	//check if in reverse
	if newState.Buttons[reverseGear] {
		c.Gear = -1
		return
	}

	//else in neutral
	c.Gear = 0
}

func (c *SmallRacerState) mapSteer(value float64, curve float64) {
	if curve < 1.0 || curve > 2.0 {
		curve = 1.0
	}

	valueWithTrim := value + c.SteerTrim

	valueWithCurve := float64(0)
	if valueWithTrim > 0 {
		valueWithCurve = vehicle.PowCurve(valueWithTrim, curve)
	} else if valueWithTrim < 0 {
		valueWithTrim = valueWithTrim * -1
		valueWithCurve = vehicle.PowCurve(valueWithTrim, curve)
		valueWithCurve = valueWithCurve * -1
	}

	c.Steer = vehicle.MapAxisWithDeadZone(valueWithCurve, MinInput, MaxInput, MinOutput, MaxOutput, DeadZone, 0)
}

func (c *SmallRacerState) mapEsc(throttle float64, brake float64) {
	throttleWithDeadzone := vehicle.MapTriggerWithDeadZone(throttle, MinInput, MaxInput, MinOutput, MaxOutput, DeadZone, -1)
	brakeWithDeadzone := vehicle.MapTriggerWithDeadZone(brake, MinInput, MaxInput, MinOutput, MaxOutput, DeadZone, -1)

	if c.Gear == 0 {
		c.Esc = 0.0
	}
	if c.Gear == -1 {
		ratio, ok := c.Ratios[c.Gear]
		if ok {
			if throttleWithDeadzone > brakeWithDeadzone {
				c.Esc = 0.0
			} else if throttleWithDeadzone < brakeWithDeadzone {
				c.Esc = vehicle.MapToRange(brakeWithDeadzone*-1, MinInput, MaxInput, ratio.Min, 0.0)
			} else {
				c.Esc = 0.0
			}
		}
	}

	if c.Gear >= 1 && c.Gear <= TopGear {
		ratio, ok := c.Ratios[c.Gear]
		if ok {
			if throttleWithDeadzone > brakeWithDeadzone {
				c.Esc = vehicle.MapToRange(throttleWithDeadzone, MinInput, MaxInput, 0.0, ratio.Max)
				//log.Printf("Throttle %.2f Final %.2f", throttle, c.Esc)
			} else if throttleWithDeadzone < brakeWithDeadzone {
				c.Esc = vehicle.MapToRange(brakeWithDeadzone*-1, MinInput, MaxInput, ratio.Min, 0.0)
			} else {
				c.Esc = 0.0
			}
		}
	}
}
