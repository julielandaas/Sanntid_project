package fsm

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/timer"
	"fmt"
	"reflect"
	"time"
)
/*
--travelTimeBetweenFloors_ms    2000
--travelTimePassingFloor_ms     500
*/




var elevator elevio.Elevator
var prev_elevator elevio.Elevator

func Fsm(port string, id string, input_buttons_fsm chan elevio.ButtonEvent, input_floors_fsm chan int, input_obstr_fsm chan bool, timer_open_door chan timer.Timer_enum,
	timer_open_door_timeout chan bool, fsm_motorDir_output chan elevio.MotorDirection, fsm_buttonLamp_output chan elevio.ButtonEvent,
	fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool, fsm_state_requests chan elevio.Elevator,
	fsm_deleteHallRequest_requests chan elevio.ButtonEvent, requests_updatedRequests_fsm chan [elevio.N_FLOORS][elevio.N_BUTTONS]bool,
	timer_detectImmobility chan timer.Timer_enum, timer_detectImmobility_timeout chan bool) {
	
	obstructed_flag := false

	// Initialize
	clearAllLights(fsm_buttonLamp_output, fsm_doorLamp_output)
	elevator = elevio.Elevator_uninitialized()
	fsm_onInitBetweenFloors(fsm_motorDir_output)

	time.Sleep(200 * time.Millisecond)
	fsm_state_requests <- elevator
	// local data
	//stop_pressed_prev := false
	//dirn_prev := elevator.Dirn
	timer_detectImmobility <- timer.Timer_stop
	timer_detectImmobility <- timer.Timer_reset

	for {
		select {
		/*
			case buttonEvent := <- input_buttons_fsm:
				//elevio.SetButtonLamp(a.Button, a.Floor, true)
				//fsm.Elevator.Requests[a.Button][a.Floor] = true
				fsm_onRequestButtonPress(buttonEvent.Floor, buttonEvent.Button, timer_open_door, timer_open_door_timeout,
					fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output)
		*/
		
		case timer_door := <-timer_open_door_timeout:
			if !obstructed_flag{
				timer_detectImmobility <- timer.Timer_stop
				timer_detectImmobility <- timer.Timer_reset
				//fmt.Printf("timer open door timeout chan\n")
				prev_elevator = elevator
				fmt.Printf("TIMER %+v\n", timer_door)
				fsm_onDoorTimeout(timer_open_door, timer_open_door_timeout,
					fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output, fsm_deleteHallRequest_requests)
	
				if !reflect.DeepEqual(prev_elevator, elevator) {
					fsm_state_requests <- elevator
				}

			}else{
				timer_open_door <- timer.Timer_stop
			}


		case new_requests := <-requests_updatedRequests_fsm:
			if elevator.Behaviour == elevio.EB_Idle{
				timer_detectImmobility <- timer.Timer_stop
				timer_detectImmobility <- timer.Timer_reset
			}
			prev_elevator = elevator
			//fmt.Printf("Recieved new requests in fsm\n")
			elevator.Requests = new_requests

			fsm_newRequests(timer_open_door, fsm_motorDir_output, fsm_buttonLamp_output, fsm_doorLamp_output, fsm_deleteHallRequest_requests)

			if !reflect.DeepEqual(prev_elevator, elevator) {
				fsm_state_requests <- elevator
			}

		case floor := <-input_floors_fsm:
			fmt.Printf("1. in input_floor_fsm\n")
			fmt.Printf("elev.beh: %+v\n obs.flag: %+v:", elevator.Behaviour, obstructed_flag)
			if elevator.Behaviour == elevio.EB_Immobile && !obstructed_flag {
				fmt.Printf("2. in input_floor_fsm\n")
				elevator.Behaviour = elevio.EB_Idle

				fsm_state_requests <- elevator
			}
			
			timer_detectImmobility <- timer.Timer_stop
			timer_detectImmobility <- timer.Timer_reset
			//fmt.Printf("floor arrival chan\n")
			prev_elevator = elevator
			//fmt.Printf("Floor (main) %+v\n", floor)
			if floor == elevio.N_FLOORS-1 || floor == 0 {
				fsm_motorDir_output <- elevio.MD_Stop
			}

			fsm_floorIndicator_output <- floor
			fsm_onFloorArrival(floor, timer_open_door, timer_open_door_timeout,
				fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output, fsm_deleteHallRequest_requests)

			if !reflect.DeepEqual(prev_elevator, elevator) {
				fsm_state_requests <- elevator
			}

		case obstructed := <-input_obstr_fsm:
			//fmt.Printf("obstructed chan\n")
			//fmt.Printf("Obstructed detected %+v\n", obstructed)
			if obstructed{
				obstructed_flag = true
			}else{
				obstructed_flag = false
			}

			if obstructed && elevator.Behaviour == elevio.EB_DoorOpen {
				
				//elevio.SetMotorDirection(elevio.MD_Stop)
				timer_open_door <- timer.Timer_stop
				//setAllLights(fsm_buttonLamp_output)

			} else if !obstructed && (elevator.Behaviour == elevio.EB_DoorOpen || elevator.Behaviour == elevio.EB_Immobile) {

				fmt.Printf("Har tatt av obstructed\n")
				//obstructed_flag = false
				elevator.Behaviour = elevio.EB_DoorOpen
				timer_open_door <- timer.Timer_stop
				timer_open_door <- timer.Timer_reset

				timer_detectImmobility <- timer.Timer_stop
				timer_detectImmobility <- timer.Timer_reset


				fsm_state_requests <- elevator
				
				//fmt.Printf("TIMER RE-STARTED Obstruction\n")
			}
		case <-timer_detectImmobility_timeout:
			
			if elevator.Behaviour == elevio.EB_Moving{
				elevator.Behaviour = elevio.EB_Immobile

				fsm_state_requests <- elevator
			}

			noRequests_flag := true
			for i := 0; i < elevio.N_FLOORS; i++ {
				for j := 0; j < elevio.N_BUTTONS; j++ {
					if elevator.Requests[i][j] == true{
						noRequests_flag = false
					}
				}
			}

			if elevator.Behaviour == elevio.EB_Idle && !noRequests_flag{
				elevator.Behaviour = elevio.EB_Immobile

				fsm_state_requests <- elevator
			}

			if elevator.Behaviour == elevio.EB_DoorOpen && obstructed_flag{
				elevator.Behaviour = elevio.EB_Immobile

				fsm_state_requests <- elevator
			}

			
			//timer_restartElevator <- timer.Timer_stop
			//timer_restartElevator <- timer.Timer_reset
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

		default:
			//nothing happens
		}
	}
}
