package main

//import "Driver-go/elevio"
import (
	"Sanntid/Driver-go/elevio"
	"fmt"
)

func main() {

	numFloors := 4


	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go elevio.Fsm_onInitBetweenFloors()


	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.Outputdevice.RequestButtonLight(a.Button, a.Floor, true)

		case a := <- drv_floors:
		  fmt.Printf("%+v\n", a)
		  if a == numFloors-1 {
		      d = elevio.MD_Down
		  } else if a == 0 {
		      d = elevio.MD_Up
		  }
		  elevio.Outputdevice.MotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.Outputdevice.MotorDirection(elevio.MD_Stop)
			} else {
				elevio.Outputdevice.MotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.Outputdevice.RequestButtonLight(b, f, false)
				}
			}
		}
	}
}
