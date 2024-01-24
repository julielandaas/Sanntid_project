package elevator

import "time"
import "sync"
import "net"
import "fmt"
import "Driver-go/elevio"

type ElevatorBehaviour int

const (
	EB_Idle   ElevatorBehaviour = 1
	EB_DoorOpen                = -1
	EB_Moving                = 0
)


type Elevator struct {
    Behavior    ElevatorBehaviour      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}

type Direction int
const { 
    D_Down  Direction = -1,
    D_Stop  = 0,
    D_Up    = 1
}

var elevator Elevator

func elevatorUninitialized() {
	elevator = {EB_Idle, -1, D_Stop}// Assuming EB_Idle is defined in Go
}


// Assuming Dirn, Behaviour, and other custom types are defined elsewhere.

// fsmOnInitBetweenFloors sets the initial state of the elevator when it's between floors.
func fsmOnInitBetweenFloors() {
	elevio.SetMotorDirection(MD_Down)
	//outputDevice.MotorDirection(D_Down) // Assuming outputDevice is a global variable or appropriately scoped in your context
	elevator.Direction = D_Down              // Assuming elevator is a global variable or appropriately scoped
	elevator.Behaviour = EB_Moving      // Assuming constants DDown and EBMoving are defined in your Go code
}

// Other necessary variables, types, and functions (like outputDevice, Elevator, Dirn, etc.) should be defined elsewhere in your codebase.

// Assuming Button, DirnBehaviourPair, and other custom types are defined elsewhere.

// fsmOnRequestButtonPress is a translation of the provided C function.
func fsmOnRequestButtonPress(btnFloor int, btnType Button) {
	fmt.Printf("\n\n%s(%d, %s)\n", "fsmOnRequestButtonPress", btnFloor, elevioButtonToString(btnType))
	//elevatorPrint(elevator)

	switch elevator.Behaviour {
	case EBDoorOpen:
		if requestsShouldClearImmediately(elevator, btnFloor, btnType) {
			timerStart(elevator.Config.DoorOpenDurationS)
		} else {
			elevator.Requests[btnFloor][btnType] = true
		}

	case EBMoving:
		elevator.Requests[btnFloor][btnType] = true

	case EBIdle:
		elevator.Requests[btnFloor][btnType] = true
		pair := requestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case EBDoorOpen:
			outputDevice.DoorLight(true)
			timerStart(elevator.Config.DoorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)

		case EBMoving:
			outputDevice.MotorDirection(elevator.Dirn)

		case EBIdle:
			// No action needed
		}
	}

	setAllLights(elevator)

	fmt.Println("\nNew state:")
	elevatorPrint(elevator)
}

// Other functions (e.g., elevioButtonToString, elevatorPrint, requestsShouldClearImmediately, etc.) are assumed to be defined elsewhere.

