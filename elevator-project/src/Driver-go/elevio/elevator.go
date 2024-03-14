package elevio


const N_FLOORS = 4
const N_BUTTONS = 3
const TRAVELTIME_BETWEENFLOORS_MS = 2000


type MotorDirection int
const (
	MD_Up   MotorDirection  = 1
	MD_Down                 = -1
	MD_Stop                 = 0
)

type ButtonType int
const (
	BT_HallUp   ButtonType  = 0
	BT_HallDown             = 1
	BT_Cab                  = 2
)

type ButtonEvent struct {
	Floor   int
	Button  ButtonType
	Value   bool
}

type Dirn int
const (
	D_Down     Dirn  = -1
	D_Stop           = 0
	D_Up             = 1
)

type ElevatorBehaviour int
const (
	EB_Idle     ElevatorBehaviour = 1	  
	EB_DoorOpen                   = -1	  
	EB_Moving                     = 0
	EB_Immobile                   = 2
)

// Skal vi ta vekk dinna?
type ClearRequestVariant int
const (
	CV_All     ClearRequestVariant = 0
	CV_InDirn                      = 1
)

type Elevator struct {
	Behaviour    ElevatorBehaviour          `json:"behaviour"`
	Floor        int                        `json:"floor"`
	Dirn         Dirn                       `json:"direction"`
	Requests     [N_FLOORS][N_BUTTONS]bool  `json:"Requests"`
	ClearRequestVariant       ClearRequestVariant
	
}


func Elevator_initialize() Elevator {
	currentFloor := GetFloor()
	
	Elevatorstate := Elevator{
		Behaviour: EB_Idle,
		Floor: currentFloor,
		Dirn: D_Stop,
		ClearRequestVariant: CV_InDirn,
		}
	return Elevatorstate
}


func Elevio_dirn_toString(d Dirn) string{
	switch d {
	case D_Up:
		return "up"
	case D_Down:
		return "down"
	case D_Stop:
		return "stop"
	default:
		return "D_UNDEFINED"
	}
}

func Elevio_behaviour_toString(b ElevatorBehaviour) string{
	switch b {
	case EB_Idle:
		return "idle"
	case EB_Moving:
		return "moving"
	case EB_DoorOpen:
		return "doorOpen"
	case EB_Immobile:
		return "immobile"
	default:
		return "EB_UNDEFINED"
	}
}


