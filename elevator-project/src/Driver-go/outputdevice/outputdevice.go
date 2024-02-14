package outputdevice

import (
	"Sanntid/Driver-go/elevio"
)




func Outputdevice(fsm_motorDir_output chan elevio.MotorDirection, fsm_buttonLamp_output chan elevio.ButtonEvent, 
	fsm_floorIndicator_output chan int, fsm_doorLamp_output chan bool){
	
	for {
		select {
		case motorDirn := <- fsm_motorDir_output:
			elevio.SetMotorDirection(motorDirn)
		
		case buttonLamp_arguments := <- fsm_buttonLamp_output:
			elevio.SetButtonLamp(buttonLamp_arguments.Button, buttonLamp_arguments.Floor, buttonLamp_arguments.Toggle)

		case floor := <- fsm_floorIndicator_output:
			elevio.SetFloorIndicator(floor)

		case doorLampTogggle := <- fsm_doorLamp_output:
			elevio.SetDoorOpenLamp(doorLampTogggle)
		/*
		case stopLampTogggle := <- fsm_stopLamp_output:
			elevio.SetStopLamp(stopLampTogggle)
			*/
		
		default:
			//nothing happens
			
		}
	
	}
}
