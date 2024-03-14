package main

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/fsm"
	"Sanntid/Driver-go/inputdevice"
	"Sanntid/Driver-go/outputdevice"
	"Sanntid/Driver-go/requests"
	"Sanntid/Driver-go/timer"
	"Sanntid/Network-go/network/main_network"
	"flag"
)


func main() {

	var port string
	var id string

	flag.StringVar(&port, "port", "", "port of elevatorserver")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	elevio.Init("localhost:"+port)
	
	input_floors_fsm := make(chan int)
	input_obstr_fsm := make(chan bool)
	input_buttons_network := make(chan elevio.ButtonEvent, 40)

	fsm_motorDirection_output := make(chan elevio.MotorDirection, 20)
	fsm_floorIndicator_output := make(chan int, 20)
	fsm_doorLamp_output := make(chan bool, 20)
	fsm_clearAllLights_output :=make(chan bool, 10)
	fsm_state_network := make(chan elevio.Elevator, 20)
	fsm_deleteHallRequest_network := make(chan elevio.ButtonEvent, 20)

	requests_hallRequests_output := make(chan [elevio.N_FLOORS][2]bool, 20)
	requests_myState_output := make(chan requests.HRAElevState, 10)
	requests_updatedRequests_fsm := make(chan [elevio.N_FLOORS][elevio.N_BUTTONS]bool, 20)
	requests_resendHallrequests_network := make(chan elevio.ButtonEvent, 20)

	network_peersList_requests := make(chan []string, 20)
	network_hallrequest_requests := make(chan elevio.ButtonEvent, 20)
	network_statesMap_requests := make(chan map[string]requests.HRAElevState, 20)
	
	timer_openDoor := make(chan timer.Timer_enum, 20)
	timer_openDoor_timeout := make(chan bool, 20)

	timer_requests := make(chan timer.Timer_enum, 20)
	timer_requests_timeout := make(chan bool, 20)

	timer_delete := make(chan timer.Timer_enum, 20)
	timer_delete_timeout := make(chan bool, 20)

	timer_states := make(chan timer.Timer_enum, 20)
	timer_states_timeout := make(chan bool, 20)

	timer_reInitCab_Ack := make(chan timer.Timer_enum, 20)
	timer_reInitCab_Ack_timeout := make(chan bool, 20)

	timer_detectImmobility := make(chan timer.Timer_enum, 20)
	timer_detectImmobility_timeout := make(chan bool, 20)


	go main_network.Main_network(id, fsm_state_network, input_buttons_network, network_hallrequest_requests, network_statesMap_requests,
		fsm_deleteHallRequest_network, timer_requests, timer_requests_timeout, timer_delete, timer_delete_timeout, timer_states, timer_states_timeout,
		network_peersList_requests, requests_resendHallrequests_network, timer_reInitCab_Ack, timer_reInitCab_Ack_timeout)
	
	go inputdevice.Inputdevice(input_buttons_network, input_floors_fsm, input_obstr_fsm)

	go outputdevice.Outputdevice(fsm_motorDirection_output, fsm_floorIndicator_output, fsm_doorLamp_output, requests_hallRequests_output, 
		requests_myState_output, fsm_clearAllLights_output)

	go fsm.Fsm(port, id, input_floors_fsm, input_obstr_fsm, timer_openDoor, timer_openDoor_timeout,
		fsm_motorDirection_output, fsm_floorIndicator_output, fsm_doorLamp_output, fsm_state_network, fsm_deleteHallRequest_network,
		requests_updatedRequests_fsm, timer_detectImmobility, timer_detectImmobility_timeout, fsm_clearAllLights_output)

	go requests.Request_assigner(id, network_hallrequest_requests, network_statesMap_requests, requests_updatedRequests_fsm,
		network_peersList_requests, requests_resendHallrequests_network, requests_hallRequests_output, requests_myState_output)
			
	go timer.Timer_openDoor(timer_openDoor, timer_openDoor_timeout)
	go timer.Timer_requests(timer_requests, timer_requests_timeout)
	go timer.Timer_deleteRequests(timer_delete, timer_delete_timeout)
	go timer.Timer_states(timer_states, timer_states_timeout)
	go timer.Timer_reAlivePeer_CabAgreement(timer_reInitCab_Ack, timer_reInitCab_Ack_timeout)
	go timer.Timer_detectImmobility(timer_detectImmobility, timer_detectImmobility_timeout)


	for {

	}
}
