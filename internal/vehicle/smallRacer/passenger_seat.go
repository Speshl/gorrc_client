package smallracer

import (
	"fmt"

	"github.com/Speshl/gorrc_client/internal/models"
	"github.com/Speshl/gorrc_client/internal/vehicle"
	"github.com/prometheus/procfs"
)

func NewPassengerSeat(seat *models.Seat) *vehicle.VehicleSeat[SmallRacerState] {
	return vehicle.NewVehicleSeat[SmallRacerState](seat, "passenger", passengerParser[SmallRacerState], passengerCenter[SmallRacerState], passengerHudUpdater[SmallRacerState])
}

func passengerParser[T SmallRacerState](oldCommand, newCommand models.ControlState, crawlerState vehicle.VehicleStateIFace[T]) vehicle.VehicleStateIFace[T] {
	newState := crawlerState.(SmallRacerState)

	return newState
}

func passengerCenter[T SmallRacerState](state vehicle.VehicleStateIFace[T]) vehicle.VehicleStateIFace[T] {
	newState := state.(SmallRacerState)
	return newState
}

func passengerHudUpdater[T SmallRacerState](state vehicle.VehicleStateIFace[T], netInfo procfs.NetDevLine) models.Hud {
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

	lines[1] = fmt.Sprintf("Esc:%.2f | Gear:%s | Steer:%.2f | Trim:%.2f",
		newState.Esc,
		newState.Ratios[newState.Gear].Name,
		newState.Steer,
		newState.SteerTrim,
	)

	return models.Hud{
		Lines: lines,
	}
}
