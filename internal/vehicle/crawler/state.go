package crawler

import (
	"github.com/Speshl/gorrc_client/internal/vehicle"
)

func (c *CrawlerState) upShift() {
	if c.Gear < TopGear {
		c.Gear++
	}
}

func (c *CrawlerState) downShift() {
	if c.Gear > -1 {
		c.Gear--
	}
}

func (c *CrawlerState) trimSteerLeft() {
	if c.SteerTrim-MaxTrimPerCycle < MinInput {
		c.SteerTrim = MinInput
	} else {
		c.SteerTrim -= MaxTrimPerCycle
	}
}

func (c *CrawlerState) trimSteerRight() {
	if c.SteerTrim+MaxTrimPerCycle > MaxInput {
		c.SteerTrim = MaxInput
	} else {
		c.SteerTrim += MaxTrimPerCycle
	}
}

func (c *CrawlerState) camCenter() {
	c.Pan = 0.0
	c.Tilt = 0.0
}

func (c *CrawlerState) turretCenter() {
	c.TurretPan = 0.0
	c.TurretTilt = 0.0
}

func (c *CrawlerState) mapSteer(value float64) {
	c.Steer = vehicle.MapAxisWithDeadZone(value+c.SteerTrim, MinInput, MaxInput, MinOutput, MaxOutput, DeadZone, 0)
}

func (c *CrawlerState) mapEsc(throttle float64, brake float64) {
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

func (c *CrawlerState) mapPan(value float64) {
	posAdjust := vehicle.MapAxisWithDeadZone(value, MinInput, MaxInput, -1*c.PanSpeed, c.PanSpeed, DeadZone, 0)

	if c.Pan+posAdjust > MaxOutput {
		c.Pan = MaxOutput
	} else if c.Pan+posAdjust < MinOutput {
		c.Pan = MinOutput
	} else {
		c.Pan += posAdjust
	}
}

func (c *CrawlerState) mapTilt(value float64) {
	posAdjust := vehicle.MapAxisWithDeadZone(value, MinInput, MaxInput, -1*c.TiltSpeed, c.TiltSpeed, DeadZone, 0)
	if c.Tilt+posAdjust > MaxOutput {
		c.Tilt = MaxOutput
	} else if c.Tilt+posAdjust < MinOutput {
		c.Tilt = MinOutput
	} else {
		c.Tilt += posAdjust
	}
}

func (c *CrawlerState) mapTrigger(value float64) {
	c.Trigger = vehicle.MapTriggerWithDeadZone(value, MinInput, MaxInput, MinOutput, MaxOutput, DeadZone, 0)
}

func (c *CrawlerState) mapTurretTilt(value float64) {
	posAdjust := vehicle.MapAxisWithDeadZone(value, MinInput, MaxInput, -1*c.TiltSpeed, c.TiltSpeed, DeadZone, 0)
	if c.TurretTilt+posAdjust > MaxOutput {
		c.TurretTilt = MaxOutput
	} else if c.TurretTilt+posAdjust < MinOutput {
		c.TurretTilt = MinOutput
	} else {
		c.TurretTilt += posAdjust
	}
}

func (c *CrawlerState) mapTurretPan(value float64) {
	posAdjust := vehicle.MapAxisWithDeadZone(value, MinInput, MaxInput, -1*c.PanSpeed, c.PanSpeed, DeadZone, 0)
	if c.TurretPan+posAdjust > MaxOutput {
		c.TurretPan = MaxOutput
	} else if c.TurretPan+posAdjust < MinOutput {
		c.TurretPan = MinOutput
	} else {
		c.TurretPan += posAdjust
	}
}
