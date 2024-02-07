package fsm

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
	"time"
	"fmt"
	
)

var Elevator elevio.Elevator

//var timer_pointer *time.Timer
//timerpointer = time.NewTimer(Elevator.Config.DoorOpenDuration_s*(time.Second))



func SetAllLights(es elevio.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, es.CabRequests[floor][btn])
		}
	}
}


func ClearAllLights(es elevio.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
	}
	elevio.SetDoorOpenLamp(false)
}


func Fsm_onInitBetweenFloors() {
	a := elevio.GetFloor()
	if a == -1 {
		elevio.SetMotorDirection(elevio.MD_Down)
		Elevator.Dirn = elevio.D_Down
		Elevator.Behaviour = elevio.EB_Moving
	Loop:
		for {
			a = elevio.GetFloor()
			if a != -1 {
				elevio.SetMotorDirection(elevio.MD_Stop)
				Elevator.Dirn = elevio.D_Stop
				Elevator.Behaviour = elevio.EB_Idle

				break Loop
			}
		}
	}
}



func Fsm_onRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, timer_p *time.Timer) {
    switch Elevator.Behaviour {
    case elevio.EB_DoorOpen:
        if requests.Requests_shouldClearImmediately(Elevator, btn_floor, btn_type) {
            //timer_pointer := time.NewTimer(Elevator.Config.DoorOpenDuration_s*(time.Second))
			//timer_start(Elevator.Config.DoorOpenDuration_s)
			timer_p.Stop()
			timer_p.Reset(Elevator.Config.DoorOpenDuration_s*time.Second)
			fmt.Printf("TIMER STARTED\n")
        } else {
            Elevator.CabRequests[btn_floor][btn_type] = true
        }

    case elevio.EB_Moving:
        Elevator.CabRequests[btn_floor][btn_type] = true

    case elevio.EB_Idle:
        Elevator.CabRequests[btn_floor][btn_type] = true
        pair := requests.Requests_chooseDirection(Elevator)
        Elevator.Dirn = pair.Dirn
        Elevator.Behaviour = pair.Behaviour
        
		switch pair.Behaviour {
        case elevio.EB_DoorOpen:
            elevio.SetDoorOpenLamp(true)
			//timer_pointer_new := time.NewTimer(Elevator.Config.DoorOpenDuration_s*(time.Second))
            //timer_start(Elevator.Config.DoorOpenDuration_s)
			timer_p.Stop()
			timer_p.Reset(Elevator.Config.DoorOpenDuration_s*time.Second)
			fmt.Printf("TIMER STARTED\n")
            Elevator = requests.Requests_clearAtCurrentFloor(Elevator)

        case elevio.EB_Moving:
            elevio.SetMotorDirection(elevio.MotorDirection(Elevator.Dirn))

        case elevio.EB_Idle:
            // No additional action required
        }
    }

    SetAllLights(Elevator)
}



func Fsm_onFloorArrival(newFloor int, timer_p *time.Timer) {
    Elevator.Floor = newFloor
	fmt.Printf("%+v\n", newFloor)

    elevio.SetFloorIndicator(Elevator.Floor)

    switch Elevator.Behaviour {
    case elevio.EB_Moving:
        if requests.Requests_shouldStop(Elevator) {
            elevio.SetMotorDirection(elevio.D_Stop)
            elevio.SetDoorOpenLamp(true)
			
			
            Elevator = requests.Requests_clearAtCurrentFloor(Elevator)
			
            //timer_start(Elevator.Config.DoorOpenDuration_s)
			timer_p.Stop()
			fmt.Printf("TIMER STARTED\n")
			timer_p.Reset(Elevator.Config.DoorOpenDuration_s*time.Second)
			
            SetAllLights(Elevator)

            Elevator.Behaviour = elevio.EB_DoorOpen
        }
    default:
        // No action required f
    }

}




func Fsm_onDoorTimeout(timer_p *time.Timer) {
    switch Elevator.Behaviour {
    case elevio.EB_DoorOpen:
        pair := requests.Requests_chooseDirection(Elevator)
        Elevator.Dirn = pair.Dirn
        Elevator.Behaviour = pair.Behaviour

        switch Elevator.Behaviour {
        case elevio.EB_DoorOpen:
            //timer_start(Elevator.Config.DoorOpenDuration_s)
			timer_p.Stop()
			timer_p.Reset(Elevator.Config.DoorOpenDuration_s*time.Second)
			fmt.Printf("TIMER STARTED\n")
            Elevator = requests.Requests_clearAtCurrentFloor(Elevator)
            SetAllLights(Elevator)
        case elevio.EB_Moving, elevio.EB_Idle:
            elevio.SetDoorOpenLamp(false)
            elevio.SetMotorDirection(elevio.MotorDirection(Elevator.Dirn))
        }

    default:
        // No additional action required
    }

}

