package main

//import "Driver-go/elevio"
import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/fsm"
	"Sanntid/Driver-go/inputdevice"
	"Sanntid/Driver-go/outputdevice"
	"Sanntid/Driver-go/timer"
)


func main() {
	
	elevio.Init("localhost:15657", elevio.N_FLOORS)

	
	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	//fsm.Fsm_onInitBetweenFloors()
	//timer_pointer_door := time.NewTimer(3*(time.Second))
	// 3 = elevator.Config.DoorOpenDuration_s
	
	//Config
	doorOpenDuration_s := 3

	//inout to fsm channels
	input_buttons_fsm := make(chan elevio.ButtonEvent)
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


	go inputdevice.Inputdevice(input_buttons_fsm, input_floors_fsm, input_obstr_fsm)
	go outputdevice.Outputdevice(fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output)
	
	go timer.Timer_handler(timer_open_door, timer_open_door_timeout, doorOpenDuration_s)
	go fsm.Fsm(input_buttons_fsm, input_floors_fsm, input_obstr_fsm, timer_open_door, timer_open_door_timeout, 
		fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output)
	
	for{

	}
}