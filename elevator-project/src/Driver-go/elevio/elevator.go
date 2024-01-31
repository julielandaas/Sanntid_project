package elevio

//import "fmt"

const N_FLOORS = 4
const N_BUTTONS = 3


var elevator Elevator
var Outputdevice ElevOutputDevice


type MotorDirection int
const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int
const (
	BT_HallUp   ButtonType = 0
	BT_HallDown             = 1
	BT_Cab                  = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type Dirn int
const (
	D_Down Dirn = -1
	D_Stop           = 0
	D_Up             = 1
)


type ElevatorBehaviour int
const (
	EB_Idle     ElevatorBehaviour = 1
	EB_DoorOpen                   = -1
	EB_Moving                     = 0
)

type ClearRequestVariant int
const (
	CV_All     ClearRequestVariant = 0
	CV_InDirn                      = 1
)

type config struct {
	clearRequestVariant ClearRequestVariant
	doorOpenDuration_s float32 
}

type Elevator struct {
	Behavior    ElevatorBehaviour `json:"behaviour"`
	Floor       int               `json:"floor"`
	Dirn   Dirn            `json:"direction"`
	CabRequests [N_FLOORS][N_BUTTONS]int          `json:"cabRequests"`
	config config
	
}


func elevator_uninitialized() Elevator {
	elevator := Elevator{
		Behavior: EB_Idle,
		Floor: -1,
		Dirn: D_Stop,
		config: config{
			clearRequestVariant: CV_InDirn,
			doorOpenDuration_s: 3.0,
			},
		}
	
	return elevator
	
}

func Fsm_onInitBetweenFloors() {
	Outputdevice.MotorDirection(MD_Down)
	elevator.Dirn = D_Down
	elevator.Behavior = EB_Moving
}

//fmt.Print("Ferdig")
func elevio_dirn_toString(d Dirn) string{
	switch d {
	case D_Up:
		return "D_Up"
	case D_Down:
		return "D_Down"
	case D_Stop:
		return "D_Stop"
	default:
		return "D_UNDEFINED"
	}
}

func elevio_button_toString(b ButtonType) string{
	switch b {
	case BT_HallUp:
		return "BT_HallUp"
	case BT_HallDown:
		return "BT_HallDown"
	case BT_Cab:
		return "BT_Cab"
	default:
		return "BT_UNDEFINED"
	}
}

