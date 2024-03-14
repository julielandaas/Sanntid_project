package requests

import (
	"Sanntid/Driver-go/elevio"
    "encoding/json"
	"fmt"
	"os/exec"
)


type HRAElevState struct {
	Behaviour   string                `json:"behaviour"`
	Floor       int                   `json:"floor"`
	Direction   string                `json:"direction"`
	CabRequests [elevio.N_FLOORS]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [elevio.N_FLOORS][2]bool `json:"hallRequests"`
	States       map[string]HRAElevState  `json:"states"`
}


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
    case elevio.D_Stop:
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
                e.Requests[e.Floor][elevio.BT_Cab]      ||
                !Requests_below(e))
    case elevio.D_Up:
        return (e.Requests[e.Floor][elevio.BT_HallUp] ||
                e.Requests[e.Floor][elevio.BT_Cab]    ||
                 !Requests_above(e))
    default:
        return true
    }
}

func Requests_shouldClearImmediately(e elevio.Elevator) bool {
    switch e.ClearRequestVariant{
    case elevio.CV_All:
        return (e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab])
    case elevio.CV_InDirn:
        return ((e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab]) && 
                ((e.Dirn == elevio.D_Up   && e.Requests[e.Floor][elevio.BT_HallUp])   ||
                (e.Dirn == elevio.D_Down && e.Requests[e.Floor][elevio.BT_HallDown])  ||
                e.Dirn == elevio.D_Stop ||
                e.Requests[e.Floor][elevio.BT_Cab]))
    default:
        return false
    }
}

func Requests_clearAtCurrentFloor_elevatoruse(e elevio.Elevator, fsm_deleteHallRequest_network chan elevio.ButtonEvent) elevio.Elevator {
    switch e.ClearRequestVariant {
    case elevio.CV_All:
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            e.Requests[e.Floor][btn] = false
            if elevio.ButtonType(btn) != elevio.BT_Cab {
                fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.ButtonType(btn), Value: false}
            }
            
        }

    case elevio.CV_InDirn:
        e.Requests[e.Floor][elevio.BT_Cab] = false
        
        switch e.Dirn {
        case elevio.D_Up:
            if !Requests_above(e) && !e.Requests[e.Floor][elevio.BT_HallUp] {
                e.Requests[e.Floor][elevio.BT_HallDown] = false
                fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown, Value: false}
            }
            e.Requests[e.Floor][elevio.BT_HallUp] = false
            fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp, Value: false}

        case elevio.D_Down:
            if !Requests_below(e) && !e.Requests[e.Floor][elevio.BT_HallDown] {
                e.Requests[e.Floor][elevio.BT_HallUp] = false
                fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp, Value: false}
            }
            e.Requests[e.Floor][elevio.BT_HallDown] = false
            fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown, Value: false}

        case elevio.D_Stop:
            e.Requests[e.Floor][elevio.BT_HallUp] = false
            fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp, Value: false}
            e.Requests[e.Floor][elevio.BT_HallDown] = false
            fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown, Value: false}
            
        default:
            e.Requests[e.Floor][elevio.BT_HallUp] = false
            fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp, Value: false}
            e.Requests[e.Floor][elevio.BT_HallDown] = false
            fsm_deleteHallRequest_network <- elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown, Value: false}
        }

    default:
    }

    return e
}

func Requests_clearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {
    switch e.ClearRequestVariant {
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



func reassign_requests(input HRAInput, id string) *[elevio.N_FLOORS][elevio.N_BUTTONS]bool {
	hraExecutable := "hall_request_assigner"
	inpuStates_mutex.Lock()
	jsonBytes, err := json.Marshal(input)
	inpuStates_mutex.Unlock()

	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return nil
	}

	ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes), "--includeCab").CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil
	}

	output := new(map[string][elevio.N_FLOORS][elevio.N_BUTTONS]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	fmt.Printf("output orders assigned: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	myRequests := (*output)[id]

	return &myRequests

}


func get_updatedRequests(input HRAInput, id string, peersList []string) [elevio.N_FLOORS][elevio.N_BUTTONS]bool {
	mycabrequests := input.States[id].CabRequests
	temp_input := HRAInput{
		HallRequests: [elevio.N_FLOORS][2]bool{{false, false}, {false, false}, {false, false}, {false, false}},
		States:       make(map[string]HRAElevState),
	}
	temp_input.HallRequests = input.HallRequests

	inpuStates_mutex.Lock()
	if len(peersList) > 1 {
		for i := 0; i < len(peersList); i++ {
			_, ok := input.States[peersList[i]]
			if ok && (input.States[peersList[i]].Behaviour != "immobile") {
				temp_input.States[peersList[i]] = input.States[peersList[i]]
			}
		}
	} else {
		temp_input.States[id] = input.States[id]
	}
	inpuStates_mutex.Unlock()

	updatedRequests := reassign_requests(temp_input, id)
	for i := 0; i < elevio.N_FLOORS; i++ {
		if mycabrequests[i] {
			updatedRequests[i][elevio.BT_Cab] = true
		}
	}
	return *updatedRequests
}
