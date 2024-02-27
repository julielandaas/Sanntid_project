package main

//import "Driver-go/elevio"
import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/fsm"
	"Sanntid/Driver-go/inputdevice"
	"Sanntid/Driver-go/outputdevice"
	"Sanntid/Driver-go/timer"
	"Sanntid/Network-go/network/main_network"
	"Sanntid/Driver-go/requests"
)


func main() {
	
	elevio.Init("localhost:15657", elevio.N_FLOORS)

	
	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	//fsm.Fsm_onInitBe floors := <
	//Config
	doorOpenDuration_s := 3

	//inout to fsm channels
	input_buttons_fsm := make(chan elevio.ButtonEvent) // FJERN DINNA ASAP
	input_floors_fsm := make(chan int)
	input_obstr_fsm := make(chan bool)
	//input_stop_fsm := make(chan bool)

	//fsm to output
	fsm_motorDir_output := make(chan elevio.MotorDirection)
	fsm_buttonLamp_output := make(chan elevio.ButtonEvent)
	fsm_floorIndicator_output := make(chan int)
	fsm_doorLamp_output := make(chan bool)
	//fsm_stopLamp_output := make(chan bool)

	//timer channel
	timer_open_door := make(chan timer.Timer_enum)
	timer_open_door_timeout := make(chan bool)

	//input_buttons_requests
	input_buttons_network := make(chan elevio.ButtonEvent)

	fsm_state_requests := make(chan elevio.Elevator)

	fsm_deleteCabRequest_requests := make(chan elevio.Elevator)

	requests_state_network := make(chan elevio.Elevator)

	go main_network.Main_network(requests_state_network, input_buttons_network)
	go inputdevice.Inputdevice(input_buttons_network, input_floors_fsm, input_obstr_fsm)
	go outputdevice.Outputdevice(fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output)
	
	go timer.Timer_handler(timer_open_door, timer_open_door_timeout, doorOpenDuration_s)
	go fsm.Fsm(input_buttons_fsm, input_floors_fsm, input_obstr_fsm, timer_open_door, timer_open_door_timeout, 
		fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output, fsm_state_requests, fsm_deleteCabRequest_requests)
	go requests.Request_assigner(fsm_state_requests, fsm_deleteCabRequest_requests, requests_state_network)
	
	for{

	}
}