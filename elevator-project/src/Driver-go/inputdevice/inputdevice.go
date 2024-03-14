package inputdevice

import (
	"Sanntid/Driver-go/elevio"
	"fmt"	
)


func Inputdevice(input_buttonEvent_network chan elevio.ButtonEvent, input_floor_fsm chan int, input_obstruction_fsm chan bool){
	
    drv_buttonEvent_input := make(chan elevio.ButtonEvent)
	drv_floor_input := make(chan int)
	drv_obstruction_input := make(chan bool)
	drv_stop_input := make(chan bool)


	go elevio.PollButtons(drv_buttonEvent_input)
	go elevio.PollFloorSensor(drv_floor_input)
	go elevio.PollObstructionSwitch(drv_obstruction_input)
	go elevio.PollStopButton(drv_stop_input)
	
	
	for {
		select {
		case button_pressed := <-drv_buttonEvent_input:
			fmt.Printf("Button is pressed %+v\n", button_pressed)
			input_buttonEvent_network <- button_pressed
		
		case floor := <- drv_floor_input:
			fmt.Printf("At floor %+v\n", floor)
			input_floor_fsm <- floor

		case obstruction := <-drv_obstruction_input:
			fmt.Printf("Obstruction detected %+v\n", obstruction)
			input_obstruction_fsm <- obstruction
			
		default:
		}
	}
}

