package fsm

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/timer"
	"fmt"
	"reflect"
	"time"
)

var elevator elevio.Elevator
var prev_elevator elevio.Elevator

func Fsm(input_buttons_fsm chan elevio.ButtonEvent, input_floors_fsm chan int, input_obstr_fsm chan bool, timer_open_door chan timer.Timer_enum,
	timer_open_door_timeout chan bool, fsm_motorDir_output chan elevio.MotorDirection, fsm_buttonLamp_output chan elevio.ButtonEvent,
	fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool, fsm_state_requests chan elevio.Elevator,
	fsm_deleteHallRequest_requests chan elevio.ButtonEvent, requests_updatedRequests_fsm chan [elevio.N_FLOORS][elevio.N_BUTTONS]bool) {

	// Initialize
	clearAllLights(fsm_buttonLamp_output, fsm_doorLamp_output)
	elevator = elevio.Elevator_uninitialized()
	fsm_onInitBetweenFloors(fsm_motorDir_output)

	time.Sleep(200 * time.Millisecond)
	fsm_state_requests <- elevator
	// local data
	//stop_pressed_prev := false
	//dirn_prev := elevator.Dirn

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
			//fmt.Printf("timer open door timeout chan\n")
			prev_elevator = elevator
			fmt.Printf("TIMER %+v\n", timer_door)
			fsm_onDoorTimeout(timer_open_door, timer_open_door_timeout,
				fsm_motorDir_output, fsm_buttonLamp_output, fsm_floorIndicator_output, fsm_doorLamp_output, fsm_deleteHallRequest_requests)

			if !reflect.DeepEqual(prev_elevator, elevator) {
				fsm_state_requests <- elevator
			}

		case new_requests := <-requests_updatedRequests_fsm:
			prev_elevator = elevator
			//fmt.Printf("Recieved new requests in fsm\n")
			elevator.Requests = new_requests

			fsm_newRequests(timer_open_door, fsm_motorDir_output, fsm_buttonLamp_output, fsm_doorLamp_output, fsm_deleteHallRequest_requests)

			if !reflect.DeepEqual(prev_elevator, elevator) {
				fsm_state_requests <- elevator
			}

		case floor := <-input_floors_fsm:
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
			if obstructed && elevator.Behaviour == elevio.EB_DoorOpen {
				//elevio.SetMotorDirection(elevio.MD_Stop)
				timer_open_door <- timer.Timer_stop
				//setAllLights(fsm_buttonLamp_output)

			} else if !obstructed && elevator.Behaviour == elevio.EB_DoorOpen {
				timer_open_door <- timer.Timer_stop
				timer_open_door <- timer.Timer_reset
				//fmt.Printf("TIMER RE-STARTED Obstruction\n")
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

		default:
			//nothing happens
		}
	}
}
