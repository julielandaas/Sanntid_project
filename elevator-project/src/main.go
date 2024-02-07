package main

//import "Driver-go/elevio"
import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/fsm"
	"time"
	"fmt"
	
)


func main() {
	fmt.Println("Started! Slay")
	//numFloors := 4
	
	elevio.Init("localhost:15657", elevio.N_FLOORS)
	fmt.Println("Her")

	fsm.Elevator = elevio.Elevator_uninitialized()
	fsm.ClearAllLights(fsm.Elevator)
	
	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	fsm.Fsm_onInitBetweenFloors()
	timer_pointer_door := time.NewTimer(3*(time.Second))
	// 3 = elevator.Config.DoorOpenDuration_s
	

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)


	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	
	
	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			//elevio.SetButtonLamp(a.Button, a.Floor, true)
			//fsm.Elevator.CabRequests[a.Button][a.Floor] = true
			fsm.Fsm_onRequestButtonPress(a.Floor, a.Button, timer_pointer_door)

			

		case a := <- drv_floors:
			fmt.Printf("Floor (main) %+v\n", a)
			elevio.SetFloorIndicator(a)
			fsm.Fsm_onFloorArrival(a, timer_pointer_door)

			/*
		  fmt.Printf("%+v\n", a)
		  if a == numFloors-1 {
		      d = elevio.MD_Down
		  } else if a == 0 {
		      d = elevio.MD_Up
		  }
		  elevio.SetMotorDirection(d)
		  */

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < elevio.N_FLOORS; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		
		case b := <- timer_pointer_door.C:
			fmt.Printf("TIMER %+v\n", b)
			fsm.Fsm_onDoorTimeout(timer_pointer_door)

		}
	}
}
