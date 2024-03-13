package fsm

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
    "Sanntid/Driver-go/timer"
	"fmt"
	"os/exec"
    "os"
    "time"
)


//var timer_pointer *time.Timer
//timerpointer = time.NewTimer(elevator.Config.DoorOpenDuration_s*(time.Second))


/*
func setAllLights(fsm_buttonLamp_output chan elevio.ButtonEvent) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            fsm_buttonLamp_output <- elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn), Toggle: elevator.Requests[floor][btn]}

			//elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.Requests[floor][btn])
		}
	}
}
*/

func clearAllLights(fsm_buttonLamp_output chan elevio.ButtonEvent, fsm_doorLamp_output chan bool) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            fsm_buttonLamp_output <- elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn), Toggle: false}

			//elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
	}

    fsm_doorLamp_output <- false
	//elevio.SetDoorOpenLamp(false)
    
}


func fsm_onInitBetweenFloors(fsm_motorDir_output chan elevio.MotorDirection) {
	//a := elevio.GetFloor()
	if elevator.Floor == -1 {
        fsm_motorDir_output <- elevio.MD_Down
		elevator.Dirn = elevio.D_Down
		elevator.Behaviour = elevio.EB_Moving
	Loop:
		for {
			a := elevio.GetFloor()
			if a != -1 {
                fsm_motorDir_output <- elevio.MD_Stop
				elevator.Dirn = elevio.D_Stop
				elevator.Behaviour = elevio.EB_Idle
                elevator.Floor = a

				break Loop
			}
		}
	}
}


/*
func fsm_onRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, timer_open_door chan timer.Timer_enum, timer_open_door_timeout chan bool, 
	fsm_motorDir_output chan elevio.MotorDirection, fsm_buttonLamp_output chan elevio.ButtonEvent, fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool, fsm_deleteHallRequest_requests chan elevio.ButtonEvent) {
    switch elevator.Behaviour {
    case elevio.EB_DoorOpen:
        if requests.Requests_shouldClearImmediately(elevator, btn_floor, btn_type) {
            //timer_pointer := time.NewTimer(elevator.Config.DoorOpenDuration_s*(time.Second))
			//timer_start(elevator.Config.DoorOpenDuration_s)
			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset
			fmt.Printf("TIMER STARTED\n")

        } else {
            elevator.Requests[btn_floor][btn_type] = true
        }

    case elevio.EB_Moving:
        elevator.Requests[btn_floor][btn_type] = true

    case elevio.EB_Idle:
        elevator.Requests[btn_floor][btn_type] = true
        pair := requests.Requests_chooseDirection(elevator)
        elevator.Dirn = pair.Dirn
        elevator.Behaviour = pair.Behaviour

        
		switch pair.Behaviour {
        case elevio.EB_DoorOpen:
            fsm_doorLamp_output <- true
			//timer_pointer_new := time.NewTimer(elevator.Config.DoorOpenDuration_s*(time.Second))
            //timer_start(elevator.Config.DoorOpenDuration_s)
			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset
			fmt.Printf("TIMER STARTED\n")
            elevator = requests.Requests_clearAtCurrentFloor(elevator, fsm_deleteHallRequest_requests)

        case elevio.EB_Moving:
            fsm_motorDir_output <- elevio.MotorDirection(elevator.Dirn)

        case elevio.EB_Idle:
            // No additional action required
        }
    }

    setAllLights(fsm_buttonLamp_output)
}
*/

func fsm_newRequests(timer_open_door chan timer.Timer_enum, fsm_motorDir_output chan elevio.MotorDirection, fsm_buttonLamp_output chan elevio.ButtonEvent, 
    fsm_doorLamp_output chan bool, fsm_deleteHallRequest_requests chan elevio.ButtonEvent) {
    switch elevator.Behaviour {

    case elevio.EB_DoorOpen:
        if requests.Requests_shouldClearImmediately(elevator) {
            //timer_pointer := time.NewTimer(elevator.Config.DoorOpenDuration_s*(time.Second))
			//timer_start(elevator.Config.DoorOpenDuration_s)

			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset
			fmt.Printf("TIMER STARTED in newRequests door open\n")

            elevator = requests.Requests_clearAtCurrentFloor_elevatoruse(elevator, fsm_deleteHallRequest_requests)
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
			//timer_pointer_new := time.NewTimer(elevator.Config.DoorOpenDuration_s*(time.Second))
            //timer_start(elevator.Config.DoorOpenDuration_s)
			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset
			fmt.Printf("TIMER STARTED in newRequests idle\n")

            elevator = requests.Requests_clearAtCurrentFloor_elevatoruse(elevator, fsm_deleteHallRequest_requests)

        case elevio.EB_Moving:
            //fmt.Printf("trying to move\n")
            fsm_motorDir_output <- elevio.MotorDirection(elevator.Dirn)

        case elevio.EB_Idle:
            // No additional action required
            //fmt.Printf("still idle\n")
        }
    }

    //setAllLights(fsm_buttonLamp_output)
}


func fsm_onFloorArrival(newFloor int,timer_open_door chan timer.Timer_enum, timer_open_door_timeout chan bool, fsm_motorDir_output chan elevio.MotorDirection, 
    fsm_buttonLamp_output chan elevio.ButtonEvent, fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool, fsm_deleteHallRequest_requests chan elevio.ButtonEvent) {
    elevator.Floor = newFloor
	fmt.Printf("%+v\n", newFloor)

    fsm_floorIndicator_output <- elevator.Floor

    switch elevator.Behaviour {
    case elevio.EB_Moving:
        if requests.Requests_shouldStop(elevator) {
            fsm_motorDir_output <- elevio.MD_Stop
            fsm_doorLamp_output <- true
			
            elevator = requests.Requests_clearAtCurrentFloor_elevatoruse(elevator, fsm_deleteHallRequest_requests)
            //fsm_deleteCabRequest_requests <- elevator
			
            //timer_start(elevator.Config.DoorOpenDuration_s)
			fmt.Printf("TIMER STARTED in on floore arrival\n")
			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset
			
            //setAllLights(fsm_buttonLamp_output)

            elevator.Behaviour = elevio.EB_DoorOpen
        }
    default:
        // No action required f
    }

}




func fsm_onDoorTimeout(timer_open_door chan timer.Timer_enum, timer_open_door_timeout chan bool, fsm_motorDir_output chan elevio.MotorDirection, 
    fsm_buttonLamp_output chan elevio.ButtonEvent, fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool, fsm_deleteHallRequest_requests chan elevio.ButtonEvent) {
    switch elevator.Behaviour {
    case elevio.EB_DoorOpen:
        pair := requests.Requests_chooseDirection(elevator)
        elevator.Dirn = pair.Dirn
        elevator.Behaviour = pair.Behaviour

        switch elevator.Behaviour {
        case elevio.EB_DoorOpen:
            //timer_start(elevator.Config.DoorOpenDuration_s)
			timer_open_door <- timer.Timer_stop
			timer_open_door <- timer.Timer_reset
			fmt.Printf("TIMER STARTED in on door timeout\n")
            elevator = requests.Requests_clearAtCurrentFloor_elevatoruse(elevator, fsm_deleteHallRequest_requests)
            //setAllLights(fsm_buttonLamp_output)

        case elevio.EB_Moving, elevio.EB_Idle:
            fsm_doorLamp_output <- false
            fsm_motorDir_output <- elevio.MotorDirection(elevator.Dirn)
        }

    default:
        // No additional action required
    }

}

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
