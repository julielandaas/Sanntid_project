package main

//import "Driver-go/elevio"
import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/fsm"
	"Sanntid/Driver-go/inputdevice"
	"Sanntid/Driver-go/outputdevice"
	"Sanntid/Driver-go/requests"
	"Sanntid/Driver-go/timer"
	"Sanntid/Network-go/network/main_network"
	"Sanntid/Restart-go/restart"
	"flag"
)

func main() {
	next_start := restart.Main_restart()
	go restart.SendUpdateToBackup(next_start)

	var port string
	flag.StringVar(&port, "port", "", "port of elevatorserver")
	flag.Parse()

	elevio.Init("localhost:"+port, elevio.N_FLOORS)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	//fsm.Fsm_onInitBe floors := <
	//Config
	doorOpenDuration_s := 3
	requests_timeout_duration_s := 1
	delete_timeout_duration_s := 3

	//input to fsm channels
	input_buttons_fsm := make(chan elevio.ButtonEvent, 30)
	input_floors_fsm := make(chan int)
	input_obstr_fsm := make(chan bool)
	//input_stop_fsm := make(chan bool)

	//fsm to output
	fsm_motorDir_output := make(chan elevio.MotorDirection, 10)
	fsm_floorIndicator_output := make(chan int, 10)
	fsm_doorLamp_output := make(chan bool, 10)
	//fsm_stopLamp_output := make(chan bool)

	//timer channel
	timer_open_door := make(chan timer.Timer_enum, 10)
	timer_open_door_timeout := make(chan bool, 10)

	timer_requests := make(chan timer.Timer_enum, 10)
	timer_requests_timeout := make(chan bool, 10)

	timer_delete := make(chan timer.Timer_enum, 10)
	timer_delete_timeout := make(chan bool, 10)

	//input_buttons_requests
	input_buttons_network := make(chan elevio.ButtonEvent, 20)

	fsm_state_requests := make(chan elevio.Elevator, 10)

	fsm_deleteHallRequest_requests := make(chan elevio.ButtonEvent, 10)

	requests_buttonLamp_output := make(chan elevio.ButtonEvent)

	requests_state_network := make(chan elevio.Elevator, 10)

	requests_updatedRequests_fsm := make(chan [elevio.N_FLOORS][elevio.N_BUTTONS]bool, 10)
	requests_deleteHallRequest_network := make(chan elevio.ButtonEvent, 10)

	network_hallrequest_requests := make(chan elevio.ButtonEvent, 10)
	network_statesMap_requests := make(chan map[string]requests.HRAElevState, 10)
	network_id_requests := make(chan string, 10)

	go main_network.Main_network(requests_state_network, input_buttons_network, network_hallrequest_requests, network_statesMap_requests, network_id_requests, 
		requests_deleteHallRequest_network, timer_requests, timer_requests_timeout, requests_timeout_duration_s, 
		timer_delete,timer_delete_timeout, delete_timeout_duration_s)
	go inputdevice.Inputdevice(input_buttons_network, input_floors_fsm, input_obstr_fsm)
	go outputdevice.Outputdevice(fsm_motorDir_output, requests_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output)

	go timer.Timer_handler(timer_open_door, timer_open_door_timeout, doorOpenDuration_s)
	go fsm.Fsm(input_buttons_fsm, input_floors_fsm, input_obstr_fsm, timer_open_door, timer_open_door_timeout,
		fsm_motorDir_output, requests_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output, fsm_state_requests, fsm_deleteHallRequest_requests,
		requests_updatedRequests_fsm)
	go requests.Request_assigner(fsm_state_requests, fsm_deleteHallRequest_requests, requests_state_network, network_hallrequest_requests,
		network_statesMap_requests, network_id_requests, requests_updatedRequests_fsm, requests_deleteHallRequest_network, requests_buttonLamp_output)
	go timer.Timer_Requests(timer_requests, timer_requests_timeout, requests_timeout_duration_s)

	go timer.Timer_deleteRequests(timer_delete,timer_delete_timeout, delete_timeout_duration_s)

	for {

	}
}
