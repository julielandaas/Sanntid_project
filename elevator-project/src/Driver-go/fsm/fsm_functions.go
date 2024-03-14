package fsm

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
    "Sanntid/Driver-go/timer"
    /*
	"fmt"
	"os/exec"
    "os"
    "time"
    */
)


func fsm_onInitBetweenFloors(fsm_motorDirection_output chan elevio.MotorDirection) {
	if elevator.Floor == -1 {
        fsm_motorDirection_output <- elevio.MD_Down
		elevator.Dirn = elevio.D_Down
		elevator.Behaviour = elevio.EB_Moving
	Loop:
		for {
			current_floor:= elevio.GetFloor()
			if current_floor != -1 {
                fsm_motorDirection_output <- elevio.MD_Stop
				elevator.Dirn = elevio.D_Stop
				elevator.Behaviour = elevio.EB_Idle
                elevator.Floor = current_floor
				break Loop
			}
		}
	}
}


func fsm_newRequests(timer_open_door chan timer.Timer_enum, fsm_motorDir_output chan elevio.MotorDirection, fsm_doorLamp_output chan bool, 
    fsm_deleteHallRequest_network chan elevio.ButtonEvent) {
    switch elevator.Behaviour {

    case elevio.EB_DoorOpen:
        if requests.Requests_shouldClearImmediately(elevator) {
			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset

            elevator = requests.Requests_clearAtCurrentFloor_elevatoruse(elevator, fsm_deleteHallRequest_network)
        }

    case elevio.EB_Moving:
        // No action required

    case elevio.EB_Idle:
        pair := requests.Requests_chooseDirection(elevator)
        elevator.Dirn = pair.Dirn
        elevator.Behaviour = pair.Behaviour

		switch pair.Behaviour {
        case elevio.EB_DoorOpen:
            fsm_doorLamp_output <- true

			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset

            elevator = requests.Requests_clearAtCurrentFloor_elevatoruse(elevator, fsm_deleteHallRequest_network)

        case elevio.EB_Moving:
            fsm_motorDir_output <- elevio.MotorDirection(elevator.Dirn)

        case elevio.EB_Idle:
            // No action required
        }
    }
}


func fsm_onFloorArrival(newFloor int,timer_open_door chan timer.Timer_enum, fsm_motorDir_output chan elevio.MotorDirection, 
    fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool, fsm_deleteHallRequest_network chan elevio.ButtonEvent) {
    
    elevator.Floor = newFloor
    fsm_floorIndicator_output <- elevator.Floor

    switch elevator.Behaviour {
    case elevio.EB_Moving:
        if requests.Requests_shouldStop(elevator) {
            fsm_motorDir_output <- elevio.MD_Stop
            fsm_doorLamp_output <- true
			
            elevator = requests.Requests_clearAtCurrentFloor_elevatoruse(elevator, fsm_deleteHallRequest_network)
            
			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset

            elevator.Behaviour = elevio.EB_DoorOpen
        }
    default:
        // No action required
    }

}


func fsm_onDoorTimeout(timer_open_door chan timer.Timer_enum, fsm_motorDirection_output chan elevio.MotorDirection, 
    fsm_doorLamp_output chan bool, fsm_deleteHallRequest_network chan elevio.ButtonEvent) {
    switch elevator.Behaviour {
  
    case elevio.EB_DoorOpen:
        pair := requests.Requests_chooseDirection(elevator)
        elevator.Dirn = pair.Dirn
        elevator.Behaviour = pair.Behaviour

        switch elevator.Behaviour {
        
        case elevio.EB_DoorOpen:
			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset

            elevator = requests.Requests_clearAtCurrentFloor_elevatoruse(elevator, fsm_deleteHallRequest_network)

        case elevio.EB_Moving, elevio.EB_Idle:
            fsm_doorLamp_output <- false
            fsm_motorDirection_output <- elevio.MotorDirection(elevator.Dirn)
        }

    default:
    }
}

/*
func restart_elevator(port string, id string){
    fmt.Println("Restarts program\n")
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "main.go", "-port", port, "-id", id)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Command finished with error: %v", err)
	}
    //panic("")
    time.Sleep(1*time.Second)

    os.Exit(0)

}
*/
