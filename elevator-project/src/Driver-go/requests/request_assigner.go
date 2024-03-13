package requests

import (
	"Sanntid/Driver-go/elevio"
	"encoding/json"
	"fmt"
	"os/exec"
	//"reflect"
	"sync"
	//"runtime"
)

var hallLightsMutex sync.Mutex

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

func setAllHallLights(all_hallrequests [elevio.N_FLOORS][2]bool, requests_buttonLamp_output chan elevio.ButtonEvent) {
	

	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS-1; btn++ {
			//hallLightsMutex.Lock()
    		//defer hallLightsMutex.Unlock()
			requests_buttonLamp_output <- elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn), Toggle: all_hallrequests[floor][btn]}

			//elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.Requests[floor][btn])
		}
	}
}

func setAllCabLights(cabrequests [elevio.N_FLOORS]bool, requests_buttonLamp_output chan elevio.ButtonEvent) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		hallLightsMutex.Lock()
		requests_buttonLamp_output <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab, Toggle: cabrequests[floor]}
		hallLightsMutex.Unlock()
		//elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.Requests[floor][btn])
	}
}

func Request_assigner(fsm_state_requests chan elevio.Elevator, fsm_deleteHallRequest_requests chan elevio.ButtonEvent, requests_state_network chan elevio.Elevator,
	network_hallrequest_requests chan elevio.ButtonEvent, network_statesMap_requests chan map[string]HRAElevState, network_id_requests chan string,
	requests_updatedRequests_fsm chan [elevio.N_FLOORS][elevio.N_BUTTONS]bool, requests_deleteHallRequest_network chan elevio.ButtonEvent, 
	requests_buttonLamp_output chan elevio.ButtonEvent, network_peersList_requests chan []string, requests_resendHallrequests_network chan elevio.ButtonEvent) {
	
		var id string
	Initialized_flag := false
	peersList := []string {}
	/*
			//HRA_output := make(chan )
		myState := HRAElevState{
			Behaviour:       "idle",
			Floor:          0,
			Direction:      "stop",
			CabRequests:    []bool{false, false, false, false}}
	*/
	input := HRAInput{
		HallRequests: [elevio.N_FLOORS][2]bool{{false, false}, {false, false}, {false, false}, {false, false}},
		States:       make(map[string]HRAElevState),
	}
	/*
	   "one": myState,
	   "two": HRAElevState{
	       Behaviour:       "idle",
	       Floor:          0,
	       Direction:      "stop",
	       CabRequests:    []bool{false, false, false, false},
	   },*/

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
		case new_peersList := <- network_peersList_requests:
			println("1. resending buttonevent \n")
			println("length Newdeadpeermap: %d length deadpeermap: %d\n", len(new_peersList), len(peersList))
			if len(new_peersList) > len(peersList){
				for i := 0; i < elevio.N_FLOORS; i++ {
					for j := 0; j < elevio.N_BUTTONS-1; j++ {
						if input.HallRequests[i][j]{
							pseudo_buttonevent:= elevio.ButtonEvent{Floor: i, Button: elevio.ButtonType(j), Toggle: true}
							println("2. resending buttonevent \n")
							requests_resendHallrequests_network <- pseudo_buttonevent
						}
						
					}
				}
			}
			peersList = new_peersList


			mycabrequests := input.States[id].CabRequests

			temp_input_states := make(map[string]HRAElevState)
			hallLightsMutex.Lock()

			if len(peersList) > 1 {
				for i := 0; i < len(peersList); i++ {
					_, ok := input.States[peersList[i]]
					if ok && ((input.States[peersList[i]].Behaviour != "immobile")) {
						temp_input_states[peersList[i]] = input.States[peersList[i]]
					}
					
				}
				input.States = temp_input_states
			}else {
				temp_input_states[id] = input.States[id]
				input.States = temp_input_states
			}

			hallLightsMutex.Unlock()
			//fmt.Printf("2. Updated input states: %+v\n", input.States)
			updatedRequests := reassign_requests(input, id)
			fmt.Printf("updated requests: %+v", updatedRequests)
			//fmt.Printf("3. Updated input states: %+v\n", input.States)
			
			for i := 0; i < elevio.N_FLOORS; i++ {
				if mycabrequests[i] {
					updatedRequests[i][elevio.BT_Cab] = true
				}
			}
			requests_updatedRequests_fsm <- *updatedRequests
			
			/*
			println("1. resending buttonevent \n")
			println("length Newdeadpeermap: %d length deadpeermap: %d\n", len(NewdeadPeerMap), len(deadPeerMap))
			if len(NewdeadPeerMap) < len(deadPeerMap){
				for i := 0; i < elevio.N_FLOORS; i++ {
					for j := 0; j < elevio.N_BUTTONS-1; j++ {
						if input.HallRequests[i][j]{
							pseudo_buttonevent:= elevio.ButtonEvent{Floor: i, Button: elevio.ButtonType(j), Toggle: true}
							println("2. resending buttonevent \n")
							requests_resendHallrequests_network <- pseudo_buttonevent
						}
						
					}
				}
			}
			deadPeerMap  = NewdeadPeerMap
			*/

			/*
			if(Initialized_flag){
				for i := 0; i < len(deadPeerLst); i++ {
					delete(input.States, deadPeerLst[i])
					
				}

				updatedRequests := reassign_requests(input, id)
				requests_updatedRequests_fsm <- *updatedRequests
			}
			*/

		case id_sent := <-network_id_requests:
			id = id_sent

		case hallRequest := <-network_hallrequest_requests:
			//fmt.Printf("Recieved on network_hallrequest_requests:%+v\n", hallRequest)

			flag_detectedUpdate := false

			switch hallRequest.Toggle {
			case true:
				if input.HallRequests[hallRequest.Floor][hallRequest.Button] != true{
					input.HallRequests[hallRequest.Floor][hallRequest.Button] = true
					flag_detectedUpdate = true
				}
			case false:
				if input.HallRequests[hallRequest.Floor][hallRequest.Button] != false{
					input.HallRequests[hallRequest.Floor][hallRequest.Button] = false
					flag_detectedUpdate = true
				}
			}

			if flag_detectedUpdate{

			setAllHallLights(input.HallRequests, requests_buttonLamp_output)

			mycabrequests := input.States[id].CabRequests

			fmt.Printf("request assigner because of new hall request\n")
			if(Initialized_flag){

				temp_input_states := make(map[string]HRAElevState)
				hallLightsMutex.Lock()

				if len(peersList) > 1 {
					for i := 0; i < len(peersList); i++ {
						_, ok := input.States[peersList[i]]
						if ok{
							temp_input_states[peersList[i]] = input.States[peersList[i]]
						}
						
					}
					input.States = temp_input_states
				}else {
					temp_input_states[id] = input.States[id]
					input.States = temp_input_states
				}
				hallLightsMutex.Unlock()

				updatedRequests := reassign_requests(input, id)

				for i := 0; i < elevio.N_FLOORS; i++ {
					if mycabrequests[i] {
						updatedRequests[i][elevio.BT_Cab] = true
					}
				}

				requests_updatedRequests_fsm <- *updatedRequests
			}
			}

			//kanskje vi skulle hatt to ulike kanaler eller noe? her sender vi kanskje fort etter hverandre

		case stateMap := <-network_statesMap_requests: 
			//fmt.Printf("Recieved state in requests\n")
			if (!Initialized_flag){
				Initialized_flag = true
			}
			/*

			// Create a copy of stateMap
			stateCopy := make(map[string]HRAElevState)
			for k, v := range stateMap {
				stateCopy[k] = v
			}*/
			// ...
			hallLightsMutex.Lock()
			// Deep-copy of states
			/*
			for k, v := range stateMap {
				input.States[k] = v
			}*/
			input.States = stateMap
			hallLightsMutex.Unlock()



			/*
			// Deep-copy of states
			for k, v := range stateMap {
				input.States[k] = v
			}
			hallLightsMutex.Unlock()
			*/
			//fmt.Printf("1. Updated input states: %+v\n", input.States)
			


			setAllHallLights(input.HallRequests, requests_buttonLamp_output)

			setAllCabLights(input.States[id].CabRequests, requests_buttonLamp_output)

			//fmt.Printf("request assigner because of new state\n")
			mycabrequests := input.States[id].CabRequests

			temp_input_states := make(map[string]HRAElevState)
			hallLightsMutex.Lock()

			if len(peersList) > 1 {
				for i := 0; i < len(peersList); i++ {
					_, ok := input.States[peersList[i]]
					if ok && ((input.States[peersList[i]].Behaviour != "immobile")) {
						temp_input_states[peersList[i]] = input.States[peersList[i]]
					}
					
				}
				input.States = temp_input_states
			}else {
				temp_input_states[id] = input.States[id]
				input.States = temp_input_states
			}

			hallLightsMutex.Unlock()
			//fmt.Printf("2. Updated input states: %+v\n", input.States)
			updatedRequests := reassign_requests(input, id)
			fmt.Printf("updated requests: %+v", updatedRequests)
			//fmt.Printf("3. Updated input states: %+v\n", input.States)
			
			for i := 0; i < elevio.N_FLOORS; i++ {
				if mycabrequests[i] {
					updatedRequests[i][elevio.BT_Cab] = true
				}
			}
			requests_updatedRequests_fsm <- *updatedRequests
			//fmt.Printf("4. Updated input states: %+v\n", input.States)
			
			
		case state := <-fsm_state_requests:
			requests_state_network <- state
			/*
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
			*/

		case delete_buttonEvent := <-fsm_deleteHallRequest_requests:
			//input.HallRequests[delete_buttonEvent.Floor][delete_buttonEvent.Button] = false
			requests_deleteHallRequest_network <- delete_buttonEvent

			//eehh fjerne dette

			// HER KAN DET VÆRE STORE PROBLEM
			/*

				switch current_state.Dirn {
				case elevio.D_Up:
					if !Requests_above(current_state) && !input.HallRequests[current_state.Floor][elevio.BT_HallUp] {
						[current_state.Floor][elevio.BT_HallDown] = false
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

			*/
		default:

		}
	}
}

func reassign_requests(input HRAInput, id string) *[elevio.N_FLOORS][elevio.N_BUTTONS]bool {
	/*
	input_temp := HRAInput{
		HallRequests: [elevio.N_FLOORS][2]bool{{false, false}, {false, false}, {false, false}, {false, false}},
		States:       make(map[string]HRAElevState)}

	// Deep copy each field
	for k, v := range input.HallRequests {
		for i, val := range v {
			input_temp.HallRequests[k][i] = val
		}
	}
	for k, v := range input.States {
		input_temp.States[k] = v
	}

	// Delete the id-s that are dead, so that they don't get assigned orders when they are dead
	for k,_ := range deadPeerMap {
		delete(input_temp.States, k)
	}
	*/
	

	
	hraExecutable := "hall_request_assigner"
	hallLightsMutex.Lock()
	jsonBytes, err := json.Marshal(input)
	hallLightsMutex.Unlock()

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
