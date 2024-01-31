package requests

import (
	"Sanntid/Driver-go/elevio"
	"fmt"
)

type DirnBehaviourPair struct {
	dirn elevio.Dirn
	behaviour elevio.ElevatorBehaviour
}

func requests_above(e elevio.Elevator) int {
    for f := e.Floor + 1; f < elevio.N_FLOORS; f++ {
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            if e.CabRequests[f][btn] == 1 {
                return 1
            }
        }
    }
    return 0
}

func requests_below(e elevio.Elevator) int {
    for f := 0; f < e.Floor; f++ {
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            if e.CabRequests[f][btn] == 1 {
                return 1
            }
        }
    }
    return 0
}

func requests_here(e elevio.Elevator) int {

    for btn := 0; btn < elevio.N_BUTTONS; btn++ {
        if e.CabRequests[e.Floor][btn] == 1 {
            return 1
        }
    }
    
    return 0
}




func requests_chooseDirection(e elevio.Elevator) DirnBehaviourPair {
    switch e.Dirn {
    case elevio.D_Up:
        if requests_above(e) == 1 {
            return DirnBehaviourPair{elevio.D_Up, elevio.EB_Moving}
        } else if requests_here(e) == 1{
            return DirnBehaviourPair{elevio.D_Down, elevio.EB_DoorOpen}
        } else if requests_below(e)  == 1{
            return DirnBehaviourPair{elevio.D_Down, elevio.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.D_Stop, elevio.EB_Idle}
        }
    case elevio.D_Down:
        if requests_below(e) == 1{
            return DirnBehaviourPair{elevio.D_Down, elevio.EB_Moving}
        } else if requests_here(e) == 1{
            return DirnBehaviourPair{elevio.D_Up, elevio.EB_DoorOpen}
        } else if requests_above(e) == 1{
            return DirnBehaviourPair{elevio.D_Up, elevio.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.D_Stop, elevio.EB_Idle}
        }
    case elevio.D_Stop: // there should only be one request in the Stop case. Checking up or down first is arbitrary.
        if requests_here(e) == 1 {
            return DirnBehaviourPair{elevio.D_Stop, elevio.EB_DoorOpen}
        } else if requests_above(e) == 1{
            return DirnBehaviourPair{elevio.D_Up, elevio.EB_Moving}
        } else if requests_below(e) == 1{
            return DirnBehaviourPair{elevio.D_Down, elevio.EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.D_Stop, elevio.EB_Idle}
        }
    default:
        return DirnBehaviourPair{elevio.D_Stop, elevio.EB_Idle}
    }
}
