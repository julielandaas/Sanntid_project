package requests

import (
	"Sanntid/Driver-go/elevio"
	"os/exec"
	"fmt"
	"encoding/json"
	"runtime"
)


type HRAElevState struct {
    Behaviour   string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}

type HRAInput struct {
    HallRequests    [][2]bool                   `json:"hallRequests"`
    States          map[string]HRAElevState     `json:"states"`
}





func Request_assigner(fsm_state_requests chan elevio.Elevator, fsm_deleteCabRequest_requests chan elevio.Elevator, requests_state_network chan elevio.Elevator){
	//HRA_output := make(chan )
	myState := HRAElevState{
		Behaviour:       "idle",
		Floor:          0,
		Direction:      "stop",
		CabRequests:    []bool{false, false, false, false}}
	
	input := HRAInput{
        HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
        States: map[string]HRAElevState{
            "one": myState,
            "two": HRAElevState{
                Behaviour:       "idle",
                Floor:          0,
                Direction:      "stop",
                CabRequests:    []bool{false, false, false, false},
            },
        },
    }

    
	for {
		select {

			//vi flytter dette over i Network og så fikser vi det her senere
			/*
		case button_request := <- input_buttons_requests:
			if button_request.Button == elevio.BT_Cab{
				myState.CabRequests[button_request.Floor] = true
				input.States["one"] = myState
			}else{
				input.HallRequests[button_request.Floor][button_request.Button] = true
			}
			reassign_requests(input)
*/
// MULIG FUCK-UP, SÅ SJEKK HER VISS DET BLIR FEIL
		case state := <- fsm_state_requests:
			requests_state_network <- state

			myState.Behaviour = elevio.Elevio_behaviour_toString(state.Behaviour)
			myState.Floor = state.Floor
			myState.Direction = elevio.Elevio_dirn_toString(state.Dirn)

			for f := 0; f < elevio.N_FLOORS; f++ {
				if state.Requests[f][elevio.BT_Cab] == true{
					myState.CabRequests[f] = true
					//input.States["one"].CabRequests[myState.Floor] = true
				}
			}

			// Put the modified struct back into the map
			input.States["one"] = myState
			reassign_requests(input)


		case current_state := <-fsm_deleteCabRequest_requests:
// HER KAN DET VÆRE STORE PROBLEM
			myState.CabRequests[current_state.Floor] = false

			switch current_state.Dirn {
			case elevio.D_Up:
				if !Requests_above(current_state) && !input.HallRequests[current_state.Floor][elevio.BT_HallUp] {
					input.HallRequests[current_state.Floor][elevio.BT_HallDown] = false
				}
				input.HallRequests[current_state.Floor][elevio.BT_HallUp] = false
	
			case elevio.D_Down:
				if !Requests_below(current_state) && !input.HallRequests[current_state.Floor][elevio.BT_HallDown] {
					input.HallRequests[current_state.Floor][elevio.BT_HallUp] = false
				}
				input.HallRequests[current_state.Floor][elevio.BT_HallDown] = false
	
			case elevio.D_Stop:
				input.HallRequests[current_state.Floor][elevio.BT_HallUp] = false
				input.HallRequests[current_state.Floor][elevio.BT_HallDown] = false
			default:
				input.HallRequests[current_state.Floor][elevio.BT_HallUp] = false
				input.HallRequests[current_state.Floor][elevio.BT_HallDown] = false
			}

			input.States["one"] = myState

		default:
			
		}

		
	}
}

func reassign_requests(input HRAInput)map[string][][2]bool{
	hraExecutable := "hall_request_assigner"
	jsonBytes, err := json.Marshal(input)
    	if err != nil {
        	fmt.Println("json.Marshal error: ", err)
        	return nil
    	}
    
    ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
    	if err != nil {
        	fmt.Println("exec.Command error: ", err)
        	fmt.Println(string(ret))
        	return nil
    	}

	output := new(map[string][][2]bool)
    	err = json.Unmarshal(ret, &output)
    	if err != nil {
        	fmt.Println("json.Unmarshal error: ", err)
        	return nil
    	}
        
    fmt.Printf("output: \n")
    	for k, v := range *output {
        	fmt.Printf("%6v :  %+v\n", k, v)
    	}

	return *output
}