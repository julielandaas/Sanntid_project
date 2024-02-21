package inputdevice

import (
	"Sanntid/Driver-go/elevio"
	"fmt"	
)




func Inputdevice(input_buttons_requests chan elevio.ButtonEvent, input_floors_fsm chan int, input_obstr_fsm chan bool){
	
    drv_buttons_input := make(chan elevio.ButtonEvent)
	drv_floors_input := make(chan int)
	drv_obstr_input := make(chan bool)
	drv_stop_input := make(chan bool)


	go elevio.PollButtons(drv_buttons_input)
	go elevio.PollFloorSensor(drv_floors_input)
	go elevio.PollObstructionSwitch(drv_obstr_input)
	go elevio.PollStopButton(drv_stop_input)
	
	
	for {
		select {
		case button_pressed := <-drv_buttons_input:
			fmt.Printf("Button is pressed %+v\n", button_pressed)
			input_buttons_requests <- button_pressed
		
		case floors := <- drv_floors_input:
			fmt.Printf("At floor %+v\n", floors)
			input_floors_fsm <- floors

		case obstructed := <-drv_obstr_input:
			fmt.Printf("Obstruction detected %+v\n", obstructed)
			input_obstr_fsm <- obstructed
		/*
		case stop_pressed := <-drv_stop_input:
			fmt.Printf("Stop pressed %+v\n", stop_pressed)
			input_stop_fsm <- stop_pressed
			*/
			
		default:
			//nothing happens
		}
	}
}

