package fsm

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/timer"
	"reflect"
	"time"
)

var elevator elevio.Elevator
var prev_elevator elevio.Elevator


func Fsm(port string, id string, input_floors_fsm chan int, input_obstr_fsm chan bool, timer_openDoor chan timer.Timer_enum, timer_openDoor_timeout chan bool, 
	fsm_motorDirection_output chan elevio.MotorDirection, fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool, fsm_state_network chan elevio.Elevator,
	fsm_deleteHallRequest_network chan elevio.ButtonEvent, requests_updatedRequests_fsm chan [elevio.N_FLOORS][elevio.N_BUTTONS]bool, timer_detectImmobility chan timer.Timer_enum, 
	timer_detectImmobility_timeout chan bool, fsm_clearAllLights_output chan bool) {
	
		
	obstructed_flag := false
	var lastKnownBehaviour_beforeImmobile elevio.ElevatorBehaviour 

	fsm_clearAllLights_output <- true
	elevator = elevio.Elevator_initialize()
	fsm_onInitBetweenFloors(fsm_motorDirection_output)

	time.Sleep(200 * time.Millisecond)
	fsm_state_network <- elevator

	timer_detectImmobility <- timer.Timer_stop
	timer_detectImmobility <- timer.Timer_reset


	for {
		select {	
		case <- timer_openDoor_timeout:
			if !obstructed_flag {
				prev_elevator = elevator
				fsm_onDoorTimeout(timer_openDoor, fsm_motorDirection_output, fsm_doorLamp_output, fsm_deleteHallRequest_network)
				
				if !reflect.DeepEqual(prev_elevator, elevator) {
					fsm_state_network <- elevator
				}
				
				timer_detectImmobility <- timer.Timer_stop
				timer_detectImmobility <- timer.Timer_reset

			} else {
				timer_openDoor <- timer.Timer_stop
			}


		case new_requests := <-requests_updatedRequests_fsm:
			if elevator.Behaviour == elevio.EB_Idle {
				timer_detectImmobility <- timer.Timer_stop
				timer_detectImmobility <- timer.Timer_reset
			}
			prev_elevator = elevator
			elevator.Requests = new_requests

			fsm_newRequests(timer_openDoor, fsm_motorDirection_output, fsm_doorLamp_output, fsm_deleteHallRequest_network)

			if !reflect.DeepEqual(prev_elevator, elevator) {
				fsm_state_network <- elevator
			}

		case floor := <-input_floors_fsm:
			if elevator.Behaviour == elevio.EB_Immobile && !obstructed_flag {
				elevator.Behaviour = lastKnownBehaviour_beforeImmobile
				fsm_state_network <- elevator
			}
			
			timer_detectImmobility <- timer.Timer_stop
			timer_detectImmobility <- timer.Timer_reset
	
			prev_elevator = elevator
			if floor == elevio.N_FLOORS-1 || floor == 0 {
				fsm_motorDirection_output <- elevio.MD_Stop
			}

			fsm_floorIndicator_output <- floor
			fsm_onFloorArrival(floor, timer_openDoor, fsm_motorDirection_output, fsm_floorIndicator_output, fsm_doorLamp_output, fsm_deleteHallRequest_network)

			if !reflect.DeepEqual(prev_elevator, elevator) {
				fsm_state_network <- elevator
			}

		case obstructed := <-input_obstr_fsm:
			if obstructed {
				obstructed_flag = true
			} else {
				obstructed_flag = false
			}

			if obstructed && elevator.Behaviour == elevio.EB_DoorOpen {
				timer_openDoor <- timer.Timer_stop

			} else if !obstructed && (elevator.Behaviour == elevio.EB_DoorOpen || elevator.Behaviour == elevio.EB_Immobile) {
				elevator.Behaviour = elevio.EB_DoorOpen
				timer_openDoor <- timer.Timer_stop
				timer_openDoor <- timer.Timer_reset

				timer_detectImmobility <- timer.Timer_stop
				timer_detectImmobility <- timer.Timer_reset

				fsm_state_network <- elevator
			}

		case <-timer_detectImmobility_timeout:
			if elevator.Behaviour == elevio.EB_Moving {
				lastKnownBehaviour_beforeImmobile = elevator.Behaviour
				elevator.Behaviour = elevio.EB_Immobile

				fsm_state_network <- elevator
			}

			noRequests_flag := true
			for i := 0; i < elevio.N_FLOORS; i++ {
				for j := 0; j < elevio.N_BUTTONS; j++ {
					if elevator.Requests[i][j] == true{
						noRequests_flag = false
					}
				}
			}

			if elevator.Behaviour == elevio.EB_Idle && !noRequests_flag {
				lastKnownBehaviour_beforeImmobile = elevator.Behaviour
				elevator.Behaviour = elevio.EB_Immobile

				fsm_state_network <- elevator
			}

			if elevator.Behaviour == elevio.EB_DoorOpen && obstructed_flag {
				lastKnownBehaviour_beforeImmobile = elevator.Behaviour
				elevator.Behaviour = elevio.EB_Immobile

				fsm_state_network <- elevator
			}

		default:
		}
	}
}
