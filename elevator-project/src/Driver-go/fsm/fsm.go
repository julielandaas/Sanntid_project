package fsm

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/timer"
	"fmt"
)

var elevator elevio.Elevator

func Fsm(input_buttons_fsm chan elevio.ButtonEvent, input_floors_fsm chan int, input_obstr_fsm chan bool, timer_open_door chan timer.Timer_enum, timer_open_door_timeout chan bool, 
	fsm_motorDir_output chan elevio.MotorDirection, fsm_buttonLamp_output chan elevio.ButtonEvent, fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool){
	
	// Initialize
	clearAllLights(fsm_buttonLamp_output, fsm_doorLamp_output)
	elevator = elevio.Elevator_uninitialized()
	fsm_onInitBetweenFloors(fsm_motorDir_output)
	
	// local data
	//stop_pressed_prev := false
	//dirn_prev := elevator.Dirn

	for {
		select {
		case buttonEvent := <- input_buttons_fsm:
			//elevio.SetButtonLamp(a.Button, a.Floor, true)
			//fsm.Elevator.CabRequests[a.Button][a.Floor] = true
			fsm_onRequestButtonPress(buttonEvent.Floor, buttonEvent.Button, timer_open_door, timer_open_door_timeout, 
				fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output)

		case floor := <- input_floors_fsm:
			//fmt.Printf("Floor (main) %+v\n", floor)
			if floor == elevio.N_FLOORS-1 || floor  == 0 {
				fsm_motorDir_output <- elevio.MD_Stop
			}

			fsm_floorIndicator_output <- floor
			fsm_onFloorArrival(floor, timer_open_door, timer_open_door_timeout, 
				fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output)

			

		case obstructed := <-input_obstr_fsm:
			//fmt.Printf("Obstructed detected %+v\n", obstructed)
			if obstructed && elevator.Behaviour == elevio.EB_DoorOpen{
				//elevio.SetMotorDirection(elevio.MD_Stop)
				timer_open_door <- timer.Timer_stop
            	setAllLights(fsm_buttonLamp_output)
				
			} else if !obstructed && elevator.Behaviour == elevio.EB_DoorOpen{
				timer_open_door <- timer.Timer_stop
				timer_open_door <- timer.Timer_reset
				fmt.Printf("TIMER RE-STARTED Obstruction\n")
			}
		/*	
		case stop_pressed := <- input_stop_fsm:
			fmt.Printf("STOP detected%+v\n", stop_pressed)
			
			if stop_pressed_prev == false && stop_pressed == true{
				stop_pressed_prev = true
				fsm_motorDir_output <- elevio.MD_Stop
				fsm_stopLamp_output <- true
				
				dirn_prev = elevator.Dirn
				elevator.Dirn = elevio.D_Stop

			} else if stop_pressed_prev == true && stop_pressed == true{
				stop_pressed_prev = false
				fsm_motorDir_output <- elevio.MotorDirection(dirn_prev)
				elevator.Dirn = dirn_prev
				fsm_stopLamp_output <- false
			}
		*/	
		
		case timer_door := <- timer_open_door_timeout:
			fmt.Printf("TIMER %+v\n", timer_door)
			fsm_onDoorTimeout(timer_open_door, timer_open_door_timeout, 
				fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output)
		
		default:
			//nothing happens
		}
	}
}