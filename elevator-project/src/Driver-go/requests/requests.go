package requests

import (
	"Sanntid/Driver-go/elevio"
	
)

type DirnBehaviourPair struct {
	Dirn      elevio.Dirn
	Behaviour elevio.ElevatorBehaviour
}

func Requests_above(e elevio.Elevator) bool {
    for f := e.Floor + 1; f < elevio.N_FLOORS; f++ {
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            if e.CabRequests[f][btn] {
                return true
            }
        }
    }
    return false
}

func Requests_below(e elevio.Elevator) bool {
    for f := 0; f < e.Floor; f++ {
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            if e.CabRequests[f][btn] {
                return true
            }
        }
    }
    return false
}

func Requests_here(e elevio.Elevator) bool {

    for btn := 0; btn < elevio.N_BUTTONS; btn++ {
        if e.CabRequests[e.Floor][btn] {
            return true
        }
    }
    
    return false
}




func Requests_chooseDirection(e elevio.Elevator) DirnBehaviourPair {
    switch e.Dirn {
    case elevio.D_Up:
        if Requests_above(e){
            return DirnBehaviourPair{elevio.D_Up, elevio.EB_Moving}
        } else if Requests_here(e){
            return DirnBehaviourPair{elevio.D_Down, elevio.EB_DoorOpen}
        } else if Requests_below(e){
            return DirnBehaviourPair{elevio.D_Down, elevio.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.D_Stop, elevio.EB_Idle}
        }
    case elevio.D_Down:
        if Requests_below(e){
            return DirnBehaviourPair{elevio.D_Down, elevio.EB_Moving}
        } else if Requests_here(e){
            return DirnBehaviourPair{elevio.D_Up, elevio.EB_DoorOpen}
        } else if Requests_above(e){
            return DirnBehaviourPair{elevio.D_Up, elevio.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.D_Stop, elevio.EB_Idle}
        }
    case elevio.D_Stop: // there should only be one request in the Stop case. Checking up or down first is arbitrary.
        if Requests_here(e){
            return DirnBehaviourPair{elevio.D_Stop, elevio.EB_DoorOpen}
        } else if Requests_above(e){
            return DirnBehaviourPair{elevio.D_Up, elevio.EB_Moving}
        } else if Requests_below(e){
            return DirnBehaviourPair{elevio.D_Down, elevio.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.D_Stop, elevio.EB_Idle}
        }
    default:
        return DirnBehaviourPair{elevio.D_Stop, elevio.EB_Idle}
    }
}

func Requests_shouldStop(e elevio.Elevator) bool {
    switch e.Dirn {
    case elevio.D_Down:
        return (e.CabRequests[e.Floor][elevio.BT_HallDown] ||
            e.CabRequests[e.Floor][elevio.BT_Cab] ||
            !Requests_below(e))
    case elevio.D_Up:
        return (e.CabRequests[e.Floor][elevio.BT_HallUp] ||
            e.CabRequests[e.Floor][elevio.BT_Cab] ||
            !Requests_above(e))
    default:
        return true
    }
}


func Requests_shouldClearImmediately(e elevio.Elevator, btn_floor int,  btn_type elevio.ButtonType) bool {
    switch e.Config.ClearRequestVariant{
    case elevio.CV_All:
        return e.Floor == btn_floor
    case elevio.CV_InDirn:
        return (e.Floor == btn_floor && e.Floor == btn_floor && (
            (e.Dirn == elevio.D_Up   && btn_type == elevio.BT_HallUp)    ||
            (e.Dirn == elevio.D_Down && btn_type == elevio.BT_HallDown)  ||
            e.Dirn == elevio.D_Stop ||
            btn_type == elevio.BT_Cab))
    default:
        return false
    }
}



func Requests_clearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {
    switch e.Config.ClearRequestVariant {
    case elevio.CV_All:
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            e.CabRequests[e.Floor][btn] = false
        }

    case elevio.CV_InDirn:
        e.CabRequests[e.Floor][elevio.BT_Cab] = false
        switch e.Dirn {
        case elevio.D_Up:
            if !Requests_above(e) && !e.CabRequests[e.Floor][elevio.BT_HallUp] {
                e.CabRequests[e.Floor][elevio.BT_HallDown] = false
            }
            e.CabRequests[e.Floor][elevio.BT_HallUp] = false

        case elevio.D_Down:
            if !Requests_below(e) && !e.CabRequests[e.Floor][elevio.BT_HallDown] {
                e.CabRequests[e.Floor][elevio.BT_HallUp] = false
            }
            e.CabRequests[e.Floor][elevio.BT_HallDown] = false

        case elevio.D_Stop:
            e.CabRequests[e.Floor][elevio.BT_HallUp] = false
            e.CabRequests[e.Floor][elevio.BT_HallDown] = false
        default:
            e.CabRequests[e.Floor][elevio.BT_HallUp] = false
            e.CabRequests[e.Floor][elevio.BT_HallDown] = false
        }

    default:
    }

    return e
}

