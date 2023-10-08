package vehicle

import (
	"context"
	"math"
)

type DriverCommand struct {
	Name  string
	Value float64
	Min   float64
	Max   float64
}

type CommandDriverIFace interface {
	Init() error
	Set(DriverCommand) error
	SetMany([]DriverCommand) error
}

type Vehicle interface {
	Init() error
	Start(context.Context) error
	//String() string
}

// Creates 32 uints each with only 1 bit. 1,2,4,8,16,32...
func BuildButtonMasks() []uint32 {
	buttonMasks := make([]uint32, 32)
	for i := 0; i < 32; i++ {
		buttonMasks[i] = uint32(math.Pow(2, float64(i)))
	}
	return buttonMasks
}
