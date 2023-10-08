package crawler

import (
	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/vehicle"
)

func NewDriverSeat(seat *models.Seat) *vehicle.VehicleSeat[CrawlerState] {
	return vehicle.NewVehicleSeat[CrawlerState](seat, driverParser[CrawlerState], driverCenter[CrawlerState])
}

func driverParser[T CrawlerState](oldCommand, newCommand models.ControlState, crawlerState vehicle.VehicleStateIFace[T]) vehicle.VehicleStateIFace[T] {
	newState := crawlerState.(CrawlerState)

	vehicle.NewPress(oldCommand, newCommand, UpShift, newState.upShift)
	vehicle.NewPress(oldCommand, newCommand, DownShift, newState.downShift)

	vehicle.NewPress(oldCommand, newCommand, TrimLeft, newState.trimSteerLeft)
	vehicle.NewPress(oldCommand, newCommand, TrimRight, newState.trimSteerRight)

	vehicle.NewPress(oldCommand, newCommand, CamCenter, newState.camCenter)

	// vehicletype.NewPress(oldCommand, newCommand, VolumeMute, newState.volumeMute)
	// vehicletype.NewPress(oldCommand, newCommand, VolumeUp, newState.volumeUp)
	// vehicletype.NewPress(oldCommand, newCommand, VolumeDown, newState.volumeDown)

	newState.mapSteer(newCommand.Axes[0])
	newState.mapEsc(newCommand.Axes[1], newCommand.Axes[2])
	newState.mapPan(newCommand.Axes[3])
	newState.mapTilt(newCommand.Axes[4])

	return newState
}

func driverCenter[T CrawlerState](state vehicle.VehicleStateIFace[T]) vehicle.VehicleStateIFace[T] {
	newState := state.(CrawlerState)
	newState.Gear = 0
	newState.Esc = 0.0
	newState.Steer = 0.0
	newState.Pan = 0.0
	newState.Tilt = 0.0
	return newState
}
