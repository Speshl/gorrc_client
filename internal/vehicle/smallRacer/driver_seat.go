package smallracer

import (
	"fmt"

	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/vehicle"
	"github.com/prometheus/procfs"
)

func NewDriverSeat(seat *models.Seat) *vehicle.VehicleSeat[SmallRacerState] {
	return vehicle.NewVehicleSeat[SmallRacerState](seat, "driver", driverParser[SmallRacerState], driverCenter[SmallRacerState], driverHudUpdater[SmallRacerState])
}

func driverParser[T SmallRacerState](oldCommand, newCommand models.ControlState, crawlerState vehicle.VehicleStateIFace[T]) vehicle.VehicleStateIFace[T] {
	newState := crawlerState.(SmallRacerState)

	vehicle.NewPress(oldCommand, newCommand, UpShift, newState.upShift)
	vehicle.NewPress(oldCommand, newCommand, DownShift, newState.downShift)
	vehicle.NewPress(oldCommand, newCommand, SwitchTransType, newState.switchTransType)

	vehicle.NewPress(oldCommand, newCommand, TrimLeft, newState.trimSteerLeft)
	vehicle.NewPress(oldCommand, newCommand, TrimRight, newState.trimSteerRight)

	newState.mapHPattern(newCommand, []int{7, 8, 9, 10, 11, 12}, 6)

	newState.mapSteer(newCommand.Axes[0], newCommand.Axes[9])
	newState.mapEsc(newCommand.Axes[1], newCommand.Axes[2])

	return newState
}

func driverCenter[T SmallRacerState](state vehicle.VehicleStateIFace[T]) vehicle.VehicleStateIFace[T] {
	newState := state.(SmallRacerState)
	newState.Gear = 0
	newState.Esc = 0.0
	newState.Steer = 0.0
	return newState
}

func driverHudUpdater[T SmallRacerState](state vehicle.VehicleStateIFace[T], netInfo procfs.NetDevLine) models.Hud {
	newState := state.(SmallRacerState)
	lines := make([]string, 2)

	lines[0] = fmt.Sprintf("RxPkt:%d | RxErr:%d | RxDrop: %d | TxPkt:%d | TxErr:%d | TxDrop: %d",
		netInfo.RxPackets,
		netInfo.RxErrors,
		netInfo.RxDropped,
		netInfo.TxPackets,
		netInfo.TxErrors,
		netInfo.TxDropped,
	)

	lines[1] = fmt.Sprintf("Esc:%.2f | Gear:%s | Type:%s | Steer:%.2f | Trim:%.2f",
		newState.Esc,
		newState.Ratios[newState.Gear].Name,
		newState.TransType,
		newState.Steer,
		newState.SteerTrim,
	)

	return models.Hud{
		Lines: lines,
	}
}
