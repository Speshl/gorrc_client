package crawler

import (
	"fmt"

	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/vehicle"
	"github.com/prometheus/procfs"
)

func NewDriverSeat(seat *models.Seat) *vehicle.VehicleSeat[CrawlerState] {
	return vehicle.NewVehicleSeat[CrawlerState](seat, "driver", driverParser[CrawlerState], driverCenter[CrawlerState], driverHudUpdater[CrawlerState])
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

func driverHudUpdater[T CrawlerState](state vehicle.VehicleStateIFace[T], netInfo procfs.NetDevLine) models.Hud {
	newState := state.(CrawlerState)
	lines := make([]string, 2)

	lines[0] = fmt.Sprintf("RxPkt:%d | RxErr:%d | RxDrop: %d | TxPkt:%d | TxErr:%d | TxDrop: %d",
		netInfo.RxPackets,
		netInfo.RxErrors,
		netInfo.RxDropped,
		netInfo.TxPackets,
		netInfo.TxErrors,
		netInfo.TxDropped,
	)

	lines[1] = fmt.Sprintf("Esc:%.2f | Gear:%s | Steer:%.2f | Trim:%.2f | Pan:%.2f | Tilt:%.2f",
		newState.Esc,
		newState.Ratios[newState.Gear].Name,
		newState.Steer,
		newState.SteerTrim,
		newState.Pan,
		newState.Tilt,
	)

	return models.Hud{
		Lines: lines,
	}
}

// func driverHudUpdater[T CrawlerState](state vehicle.VehicleStateIFace[T], netInfo procfs.NetDevLine) models.Hud {
// 	newState := state.(CrawlerState)
// 	lines := make([]string, 12)
// 	lines[0] = fmt.Sprintf("Esc:%.2f", newState.Esc)
// 	lines[1] = fmt.Sprintf("Gear:%s", newState.Ratios[newState.Gear].Name)
// 	lines[2] = fmt.Sprintf("Steer:%.2f", newState.Steer)
// 	lines[3] = fmt.Sprintf("Trim:%.2f", newState.SteerTrim)
// 	lines[4] = fmt.Sprintf("Pan:%.2f", newState.Pan)
// 	lines[5] = fmt.Sprintf("Tilt:%.2f", newState.Tilt)

// 	lines[6] = fmt.Sprintf("RxPkt:%d", netInfo.RxPackets)
// 	lines[7] = fmt.Sprintf("RxErr:%d", netInfo.RxErrors)
// 	lines[8] = fmt.Sprintf("RxDrop: %d", netInfo.RxDropped)

// 	lines[9] = fmt.Sprintf("TxPkt:%d", netInfo.TxPackets)
// 	lines[10] = fmt.Sprintf("TxErr:%d", netInfo.TxErrors)
// 	lines[11] = fmt.Sprintf("TxDrop: %d", netInfo.TxDropped)

// 	return models.Hud{
// 		Lines: lines,
// 	}
// }
