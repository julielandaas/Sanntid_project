package elevio

import "time"


const N_FLOORS = 4
const N_BUTTONS = 3


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
	Toggle bool
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

type Config struct {
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s  time.Duration
}

type Elevator struct {
	Behaviour    ElevatorBehaviour          `json:"behaviour"`
	Floor        int                        `json:"floor"`
	Dirn         Dirn                       `json:"direction"`
	Requests     [N_FLOORS][N_BUTTONS]bool  `json:"Requests"`
	Config       Config
	
}


func Elevator_uninitialized() Elevator {
	Elevatorstate := Elevator{
		Behaviour: EB_Idle,
		Floor: -1,
		Dirn: D_Stop,
		Config: Config{
			ClearRequestVariant: CV_InDirn,
			DoorOpenDuration_s: 3.0,
			},
		}
	return Elevatorstate
}



//fmt.Print("Ferdig")
func Elevio_dirn_toString(d Dirn) string{
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

func Elevio_behaviour_toString(b ElevatorBehaviour) string{
	switch b {
	case EB_Idle:
		return "EB_Idle"
	case EB_Moving:
		return "EB_Moving"
	case EB_DoorOpen:
		return "EB_DoorOpen"
	default:
		return "EB_UNDEFINED"
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

