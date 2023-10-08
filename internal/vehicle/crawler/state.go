package crawler

import (
	"log"

	"github.com/Speshl/gorrc_client/internal/vehicle"
)

func (c *CrawlerState) upShift() {
	log.Println("up shift")
	if c.Gear < TopGear {
		c.Gear++
	}
}

func (c *CrawlerState) downShift() {
	log.Println("down shift")
	if c.Gear > -1 {
		c.Gear--
	}
}

func (c *CrawlerState) trimSteerLeft() {
	log.Println("trim steer left")
	if c.SteerTrim-MaxTrimPerCycle < MinInput {
		c.SteerTrim = MinInput
	} else {
		c.SteerTrim -= MaxTrimPerCycle
	}
}

func (c *CrawlerState) trimSteerRight() {
	log.Println("trim steer right")
	if c.SteerTrim+MaxTrimPerCycle > MaxInput {
		c.SteerTrim = MaxInput
	} else {
		c.SteerTrim += MaxTrimPerCycle
	}
}

func (c *CrawlerState) camCenter() {
	log.Println("cam center")
	c.Pan = 0.0
	c.Tilt = 0.0
}

func (c *CrawlerState) turretCenter() {
	log.Println("turret center")
	c.TurretPan = 0.0
	c.TurretTilt = 0.0
}

// func (c *CrawlerState) volumeMute() {
// 	log.Println("volume mute")
// 	c.Volume = MinVolume
// }

// func (c *CrawlerState) volumeUp() {
// 	log.Println("volume up")
// 	if c.Volume+MaxVolumePerCycle > MaxVolume {
// 		c.Volume = MaxVolume
// 	} else {
// 		c.Volume += MaxVolumePerCycle
// 	}
// }

// func (c *CrawlerState) volumeDown() {
// 	log.Println("volume down")
// 	if c.Volume-MaxVolumePerCycle < MinVolume {
// 		c.Volume = MinVolume
// 	} else {
// 		c.Volume -= MaxVolumePerCycle
// 	}
// }

func (c *CrawlerState) mapSteer(value float64) {
	value = vehicle.GetValueWithMidDeadZone(value, 0, DeadZone)
	c.Steer = vehicle.MapToRange(value+c.SteerTrim, MinInput, MaxInput, MinOutput, MaxOutput)
}

func (c *CrawlerState) mapEsc(throttle float64, brake float64) {
	throttle = vehicle.GetValueWithLowDeadZone(throttle, 0, DeadZone)
	brake = vehicle.GetValueWithLowDeadZone(brake, 0, DeadZone)

	if c.Gear == 0 {
		c.Esc = 0.0
	}
	if c.Gear == -1 {
		ratio, ok := c.Ratios[c.Gear]
		if ok {
			if throttle > brake {
				c.Esc = 0.0
			} else if throttle < brake {
				c.Esc = vehicle.MapToRange(brake*-1, MinInput, MaxInput, ratio.Min, 0.0)
			} else {
				c.Esc = 0.0
			}
		}
	}

	if c.Gear >= 1 && c.Gear <= TopGear {
		ratio, ok := c.Ratios[c.Gear]
		if ok {
			if throttle > brake {
				c.Esc = vehicle.MapToRange(throttle, MinInput, MaxInput, 0.0, ratio.Max)
			} else if throttle < brake {
				c.Esc = vehicle.MapToRange(brake*-1, MinInput, MaxInput, ratio.Min, 0.0)
			} else {
				c.Esc = 0.0
			}
		}
	}
}

func (c *CrawlerState) mapPan(value float64) {
	value = vehicle.GetValueWithMidDeadZone(value, 0, DeadZone)

	posAdjust := vehicle.MapToRange(value, MinInput, MaxInput, -1*MaxPanPerCycle, MaxPanPerCycle)
	if c.Pan+posAdjust > MaxOutput {
		c.Pan = MaxOutput
	} else if c.Pan+posAdjust < MinOutput {
		c.Pan = MinOutput
	} else {
		c.Pan += posAdjust
	}
}

func (c *CrawlerState) mapTilt(value float64) {
	value = vehicle.GetValueWithMidDeadZone(value, 0, DeadZone)

	posAdjust := vehicle.MapToRange(value, MinInput, MaxInput, -1*MaxTiltPerCycle, MaxTiltPerCycle)
	if c.Tilt+posAdjust > MaxOutput {
		c.Tilt = MaxOutput
	} else if c.Tilt+posAdjust < MinOutput {
		c.Tilt = MinOutput
	} else {
		c.Tilt += posAdjust
	}
}

func (c *CrawlerState) mapTrigger(value float64) {
	value = vehicle.GetValueWithMidDeadZone(value, 0, DeadZone)
	c.Steer = vehicle.MapToRange(value, MinInput, MaxInput, MinOutput, MaxOutput)
}

func (c *CrawlerState) mapTurretTilt(value float64) {
	value = vehicle.GetValueWithMidDeadZone(value, 0, DeadZone)

	posAdjust := vehicle.MapToRange(value, MinInput, MaxInput, -1*MaxTiltPerCycle, MaxTiltPerCycle)
	if c.TurretTilt+posAdjust > MaxOutput {
		c.TurretTilt = MaxOutput
	} else if c.TurretTilt+posAdjust < MinOutput {
		c.TurretTilt = MinOutput
	} else {
		c.TurretTilt += posAdjust
	}
}

func (c *CrawlerState) mapTurretPan(value float64) {
	value = vehicle.GetValueWithMidDeadZone(value, 0, DeadZone)

	posAdjust := vehicle.MapToRange(value, MinInput, MaxInput, -1*MaxTiltPerCycle, MaxTiltPerCycle)
	if c.TurretPan+posAdjust > MaxOutput {
		c.TurretPan = MaxOutput
	} else if c.TurretPan+posAdjust < MinOutput {
		c.TurretPan = MinOutput
	} else {
		c.TurretPan += posAdjust
	}
}
