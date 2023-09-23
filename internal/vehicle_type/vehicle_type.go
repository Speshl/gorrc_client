package vehicleType

import (
	"context"
	"math"

	"github.com/Speshl/gorrc_client/internal/models"
)

type VehicleType interface {
	Start(ctx context.Context) error
	String() string
}

func MapToRange(value, min, max, minReturn, maxReturn float64) float64 {
	mappedValue := (maxReturn-minReturn)*(value-min)/(max-min) + minReturn

	if mappedValue > maxReturn {
		return maxReturn
	} else if mappedValue < minReturn {
		return minReturn
	} else {
		return mappedValue
	}
}

func ParseButtons(bitButton uint32, masks []uint32) []bool {
	returnvalue := make([]bool, 32)
	for i := range masks {
		returnvalue[i] = ((bitButton & masks[i]) != 0) //Check if bitbutton and mask both have bits in same place
	}
	return returnvalue
}

// Creates 32 uints each with only 1 bit. 1,2,4,8,16,32...
func BuildButtonMasks() []uint32 {
	buttonMasks := make([]uint32, 32)
	for i := 0; i < 32; i++ {
		buttonMasks[i] = uint32(math.Pow(2, float64(i)))
	}
	return buttonMasks
}

func NewPress(oldState, newState models.ControlState, buttonIndex int, f func()) {
	if newState.Buttons[buttonIndex] && !oldState.Buttons[buttonIndex] {
		f()
	}
}

func GetValueWithMidDeadZone(value, midValue, deadZone float64) float64 {
	if value > midValue && midValue+deadZone > value {
		return midValue
	} else if value < midValue && midValue-deadZone < value {
		return midValue
	}
	return value
}

func GetValueWithLowDeadZone(value, lowValue, deadZone float64) float64 {
	if value > lowValue && lowValue+deadZone > value {
		return lowValue
	}
	return value
}
