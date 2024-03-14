package requests

import (
	"Sanntid/Driver-go/elevio"
	"sync"
)

var inpuStates_mutex sync.Mutex

func Request_assigner(id string, network_hallrequest_requests chan elevio.ButtonEvent, network_statesMap_requests chan map[string]HRAElevState, 
	requests_updatedRequests_fsm chan [elevio.N_FLOORS][elevio.N_BUTTONS]bool,
	network_peersList_requests chan []string, requests_resendHallrequests_network chan elevio.ButtonEvent, 
	requests_hallRequests_output chan [elevio.N_FLOORS][2]bool, requests_myState_output chan HRAElevState) {

	peersList := []string{}

	input := HRAInput{
		HallRequests: [elevio.N_FLOORS][2]bool{{false, false}, {false, false}, {false, false}, {false, false}},
		States:       make(map[string]HRAElevState),
	}

	for {
		select {
		case new_peersList := <-network_peersList_requests:
			if len(new_peersList) > len(peersList) {
				for floor := 0; floor < elevio.N_FLOORS; floor++ {
					for button := 0; button < elevio.N_BUTTONS-1; button++ {
						if input.HallRequests[floor][button] {
							pseudo_buttonevent := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button), Value: true}
							requests_resendHallrequests_network <- pseudo_buttonevent
						}
					}
				}
			}
			peersList = new_peersList
			if len(input.States) > 0 {
				updatedRequests := get_updatedRequests(input, id, peersList)
				requests_updatedRequests_fsm <- updatedRequests
			}

		case hallRequest := <-network_hallrequest_requests:
			flag_detectedUpdate := false

			switch hallRequest.Value {
			case true:
				if !input.HallRequests[hallRequest.Floor][hallRequest.Button] {
					input.HallRequests[hallRequest.Floor][hallRequest.Button] = true
					flag_detectedUpdate = true
				}
			case false:
				if input.HallRequests[hallRequest.Floor][hallRequest.Button] {
					input.HallRequests[hallRequest.Floor][hallRequest.Button] = false
					flag_detectedUpdate = true
				}
			}

			if flag_detectedUpdate {
				requests_hallRequests_output <- input.HallRequests
				if len(input.States) > 0 {
					updatedRequests := get_updatedRequests(input, id, peersList)
					requests_updatedRequests_fsm <- updatedRequests
				}
			}

		case stateMap := <-network_statesMap_requests:
			inpuStates_mutex.Lock()
			input.States = stateMap
			inpuStates_mutex.Unlock()

			requests_hallRequests_output <- input.HallRequests
			requests_myState_output <- input.States[id]

			if len(input.States) > 0 {
				updatedRequests := get_updatedRequests(input, id, peersList)
				requests_updatedRequests_fsm <- updatedRequests
			}

		default:
		}
	}
}

