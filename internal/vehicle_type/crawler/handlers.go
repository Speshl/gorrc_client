package crawler

import (
	"log"

	vehicleType "github.com/Speshl/gorrc_client/internal/vehicle_type"
)

func (c *Crawler) upShift() {
	log.Println("up shift")
	if c.Gear < TopGear {
		c.Gear++
	}
}

func (c *Crawler) downShift() {
	log.Println("down shift")
	if c.Gear > -1 {
		c.Gear--
	}
}

func (c *Crawler) trimLeft() {
	log.Println("trim left")
	if c.SteerTrim-MaxTrimPerCycle < MinInput {
		c.SteerTrim = MinInput
	} else {
		c.SteerTrim -= MaxTrimPerCycle
	}
}

func (c *Crawler) trimRight() {
	log.Println("trim right")
	if c.SteerTrim+MaxTrimPerCycle > MaxInput {
		c.SteerTrim = MaxInput
	} else {
		c.SteerTrim += MaxTrimPerCycle
	}
}

func (c *Crawler) camCenter() {
	log.Println("cam center")
	c.Pan = 0.0
	c.Tilt = 0.0
}

func (c *Crawler) volumeMute() {
	log.Println("volume mute")
	c.Volume = MinVolume
}

func (c *Crawler) volumeUp() {
	log.Println("volume up")
	if c.Volume+MaxVolumePerCycle > MaxVolume {
		c.Volume = MaxVolume
	} else {
		c.Volume += MaxVolumePerCycle
	}
}

func (c *Crawler) volumeDown() {
	log.Println("volume down")
	if c.Volume-MaxVolumePerCycle < MinVolume {
		c.Volume = MinVolume
	} else {
		c.Volume -= MaxVolumePerCycle
	}
}

func (c *Crawler) mapSteer(value float64) {
	value = vehicleType.GetValueWithMidDeadZone(value, 0, DeadZone)
	c.Steer = vehicleType.MapToRange(value+c.SteerTrim, MinInput, MaxInput, MinOutput, MaxOutput)
}

func (c *Crawler) mapEsc(throttle float64, brake float64) {
	throttle = vehicleType.GetValueWithLowDeadZone(throttle, 0, DeadZone)
	brake = vehicleType.GetValueWithLowDeadZone(brake, 0, DeadZone)

	if c.Gear == 0 {
		c.Esc = 0.0
	}
	if c.Gear == -1 {
		ratio, ok := c.Ratios[c.Gear]
		if ok {
			if throttle > brake {
				c.Esc = 0.0
			} else if throttle < brake {
				c.Esc = vehicleType.MapToRange(brake*-1, MinInput, MaxInput, ratio.Min, 0.0)
			} else {
				c.Esc = 0.0
			}
		}
	}

	if c.Gear >= 1 && c.Gear <= TopGear {
		ratio, ok := c.Ratios[c.Gear]
		if ok {
			if throttle > brake {
				c.Esc = vehicleType.MapToRange(throttle, MinInput, MaxInput, 0.0, ratio.Max)
			} else if throttle < brake {
				c.Esc = vehicleType.MapToRange(brake*-1, MinInput, MaxInput, ratio.Min, 0.0)
			} else {
				c.Esc = 0.0
			}
		}
	}
}

func (c *Crawler) mapPan(value float64) {
	value = vehicleType.GetValueWithMidDeadZone(value, 0, DeadZone)

	posAdjust := vehicleType.MapToRange(value, MinInput, MaxInput, -1*MaxPanPerCycle, MaxPanPerCycle)
	if c.Pan+posAdjust > MaxOutput {
		c.Pan = MaxOutput
	} else if c.Pan+posAdjust < MinOutput {
		c.Pan = MinOutput
	} else {
		c.Pan += posAdjust
	}
}

func (c *Crawler) mapTilt(value float64) {
	value = vehicleType.GetValueWithMidDeadZone(value, 0, DeadZone)

	posAdjust := vehicleType.MapToRange(value, MinInput, MaxInput, -1*MaxTiltPerCycle, MaxTiltPerCycle)
	if c.Tilt+posAdjust > MaxOutput {
		c.Tilt = MaxOutput
	} else if c.Tilt+posAdjust < MinOutput {
		c.Tilt = MinOutput
	} else {
		c.Tilt += posAdjust
	}
}
