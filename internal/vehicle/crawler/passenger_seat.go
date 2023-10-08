package crawler

import (
	"fmt"

	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/vehicle"
)

func NewPassengerSeat(seat *models.Seat) *vehicle.VehicleSeat[CrawlerState] {
	return vehicle.NewVehicleSeat[CrawlerState](seat, passengerParser[CrawlerState], passngerCenter[CrawlerState], passengerHudUpdater[CrawlerState])
}

func passengerParser[T CrawlerState](oldCommand, newCommand models.ControlState, crawlerState vehicle.VehicleStateIFace[T]) vehicle.VehicleStateIFace[T] {
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

func passngerCenter[T CrawlerState](state vehicle.VehicleStateIFace[T]) vehicle.VehicleStateIFace[T] {
	newState := state.(CrawlerState)
	newState.Trigger = 0.0
	newState.TurretPan = 0.0
	newState.TurretTilt = 0.0
	return newState
}

func passengerHudUpdater[T CrawlerState](state vehicle.VehicleStateIFace[T]) models.Hud {
	newState := state.(CrawlerState)
	lines := make([]string, 3)
	lines[0] = fmt.Sprintf("Trigger: %.2f", newState.Esc)
	lines[1] = fmt.Sprintf("TurretPan: %.2f", newState.TurretPan)
	lines[2] = fmt.Sprintf("TurretTilt: %.2f", newState.TurretTilt)

	return models.Hud{
		Lines: lines,
	}
}
