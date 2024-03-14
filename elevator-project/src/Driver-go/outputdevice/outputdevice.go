package outputdevice

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
)

func setAllHallLights(all_hallrequests [elevio.N_FLOORS][2]bool) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS-1; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, all_hallrequests[floor][btn])
		}
	}
}

func setAllCabLights(cabrequests [elevio.N_FLOORS]bool) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, cabrequests[floor])
	}
}

func clearAllLights() {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
	}
	elevio.SetDoorOpenLamp(false)    
}


func Outputdevice(fsm_motorDirection_output chan elevio.MotorDirection,	fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool, requests_hallRequests_output chan [elevio.N_FLOORS][2]bool,
	requests_myState_output chan requests.HRAElevState, fsm_clearAllLights_output chan bool){
	
	for {
		select {
		case motorDirn := <- fsm_motorDirection_output:
			elevio.SetMotorDirection(motorDirn)
		
		case <- fsm_clearAllLights_output:
			clearAllLights()

		case floor := <- fsm_floorIndicator_output:
			elevio.SetFloorIndicator(floor)

		case doorLampValue := <- fsm_doorLamp_output:
			elevio.SetDoorOpenLamp(doorLampValue)

		case all_hallRequests := <- requests_hallRequests_output:
			setAllHallLights(all_hallRequests)

		case myState := <- requests_myState_output:
			setAllCabLights(myState.CabRequests) 
	
		default:
		}
	
	}
}


