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
            if e.Requests[f][btn] {
                return true
            }
        }
    }
    return false
}

func Requests_below(e elevio.Elevator) bool {
    for f := 0; f < e.Floor; f++ {
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            if e.Requests[f][btn] {
                return true
            }
        }
    }
    return false
}

func Requests_here(e elevio.Elevator) bool {
    for btn := 0; btn < elevio.N_BUTTONS; btn++ {
        if e.Requests[e.Floor][btn] {
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
        return (e.Requests[e.Floor][elevio.BT_HallDown] ||
            e.Requests[e.Floor][elevio.BT_Cab] ||
            !Requests_below(e))
    case elevio.D_Up:
        return (e.Requests[e.Floor][elevio.BT_HallUp] ||
            e.Requests[e.Floor][elevio.BT_Cab] ||
            !Requests_above(e))
    default:
        return true
    }
}


func Requests_shouldClearImmediately(e elevio.Elevator) bool {
    switch e.Config.ClearRequestVariant{
    case elevio.CV_All:
        return (e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab])
    case elevio.CV_InDirn:
        return ((e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab]) && 
                ((e.Dirn == elevio.D_Up   && e.Requests[e.Floor][elevio.BT_HallUp])    ||
                (e.Dirn == elevio.D_Down && e.Requests[e.Floor][elevio.BT_HallDown])  ||
                e.Dirn == elevio.D_Stop ||
                e.Requests[e.Floor][elevio.BT_Cab]))
    default:
        return false
    }
}



func Requests_clearAtCurrentFloor_elevatoruse(e elevio.Elevator, fsm_deleteHallRequest_requests chan elevio.ButtonEvent) elevio.Elevator {
    switch e.Config.ClearRequestVariant {
    case elevio.CV_All:
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            e.Requests[e.Floor][btn] = false
            if elevio.ButtonType(btn) != elevio.BT_Cab {
                fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.ButtonType(btn), Toggle: false}
            }
            
        }

    case elevio.CV_InDirn:
        e.Requests[e.Floor][elevio.BT_Cab] = false
        
        switch e.Dirn {
        case elevio.D_Up:
            if !Requests_above(e) && !e.Requests[e.Floor][elevio.BT_HallUp] {
                e.Requests[e.Floor][elevio.BT_HallDown] = false
                fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown, Toggle: false}
            }
            e.Requests[e.Floor][elevio.BT_HallUp] = false
            fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp, Toggle: false}

        case elevio.D_Down:
            if !Requests_below(e) && !e.Requests[e.Floor][elevio.BT_HallDown] {
                e.Requests[e.Floor][elevio.BT_HallUp] = false
                fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp, Toggle: false}
            }
            e.Requests[e.Floor][elevio.BT_HallDown] = false
            fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown, Toggle: false}

        case elevio.D_Stop:
            e.Requests[e.Floor][elevio.BT_HallUp] = false
            fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp, Toggle: false}
            e.Requests[e.Floor][elevio.BT_HallDown] = false
            fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown, Toggle: false}
            
        default:
            e.Requests[e.Floor][elevio.BT_HallUp] = false
            fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp, Toggle: false}
            e.Requests[e.Floor][elevio.BT_HallDown] = false
            fsm_deleteHallRequest_requests <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown, Toggle: false}
        }

    default:
    }

    return e
}

func Requests_clearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {
    switch e.Config.ClearRequestVariant {
    case elevio.CV_All:
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            e.Requests[e.Floor][btn] = false
        }

    case elevio.CV_InDirn:
        e.Requests[e.Floor][elevio.BT_Cab] = false
        
        switch e.Dirn {
        case elevio.D_Up:
            if !Requests_above(e) && !e.Requests[e.Floor][elevio.BT_HallUp] {
                e.Requests[e.Floor][elevio.BT_HallDown] = false
            }
            e.Requests[e.Floor][elevio.BT_HallUp] = false
        case elevio.D_Down:
            if !Requests_below(e) && !e.Requests[e.Floor][elevio.BT_HallDown] {
                e.Requests[e.Floor][elevio.BT_HallUp] = false
            }
            e.Requests[e.Floor][elevio.BT_HallDown] = false
        case elevio.D_Stop:
            e.Requests[e.Floor][elevio.BT_HallUp] = false
            e.Requests[e.Floor][elevio.BT_HallDown] = false
        default:
            e.Requests[e.Floor][elevio.BT_HallUp] = false
            e.Requests[e.Floor][elevio.BT_HallDown] = false
        }

    default:
    }

    return e
}
