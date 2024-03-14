package timer

import (
	"Sanntid/Driver-go/elevio"
	"time"
)

const OPENDOOR_TIMEOUT_DURATION_S = 3
const REQUESTS_TIMEOUT_DURATION_MS = 50
const DELETE_TIMEOUT_DURATION_MS = 60
const STATES_TIMEOUT_DURATION_MS = 50
const REALIVE_PEER_CABAGREEMENT_TIMEOUT_MS = 300
const DETECT_IMMOBILITY_TIMEOUT_DURATION_MS = elevio.TRAVELTIME_BETWEENFLOORS_MS*2


type Timer_enum int
const (
	Timer_stop     Timer_enum = 1
	Timer_reset               = -1  
)


func Timer_openDoor(timer_openDoor chan Timer_enum, timer_openDoor_timeout chan bool){
	timer_pointerDoor := time.NewTimer(time.Duration(OPENDOOR_TIMEOUT_DURATION_S)*(time.Second))
	timer_pointerDoor.Stop()

	for{
		select{
		case timer_info := <- timer_openDoor:
			switch(timer_info){
			case Timer_stop:
				timer_pointerDoor.Stop()
			case Timer_reset:
				timer_pointerDoor.Reset(time.Duration(OPENDOOR_TIMEOUT_DURATION_S)*(time.Second))
			}

		case <- timer_pointerDoor.C:
			timer_openDoor_timeout <- true

		default:
		}

	}
}

func Timer_requests(timer_requests chan Timer_enum,timer_requests_timeout chan bool){
	timer_pointer_requests := time.NewTimer(time.Duration(REQUESTS_TIMEOUT_DURATION_MS)*(time.Millisecond))
	timer_pointer_requests.Stop()

	for{
		select{
		case timer_info := <- timer_requests:
			switch(timer_info){
			case Timer_stop:
				timer_pointer_requests.Stop()
			case Timer_reset:
				timer_pointer_requests.Reset(time.Duration(REQUESTS_TIMEOUT_DURATION_MS)*(time.Millisecond))
			}

		case <- timer_pointer_requests.C:
			timer_requests_timeout <- true

		default:
		}

	}
}

func Timer_deleteRequests(timer_delete chan Timer_enum,timer_delete_timeout chan bool){
	timer_pointer_delete := time.NewTimer(time.Duration(DELETE_TIMEOUT_DURATION_MS)*(time.Millisecond))
	timer_pointer_delete.Stop()

	for{
		select{
		case timer_info := <- timer_delete:
			switch(timer_info){
			case Timer_stop:
				timer_pointer_delete.Stop()
			case Timer_reset:
			timer_pointer_delete.Reset(time.Duration(DELETE_TIMEOUT_DURATION_MS)*(time.Millisecond))
			}

		case <- timer_pointer_delete.C:
			timer_delete_timeout <- true

		default:
		}

	}
}

func Timer_states(timer_states chan Timer_enum,timer_states_timeout chan bool){
	timer_pointer_states := time.NewTimer(time.Duration(STATES_TIMEOUT_DURATION_MS)*(time.Millisecond))
	timer_pointer_states.Stop()

	for{
		select{
		case timer_info := <- timer_states:
			switch(timer_info){
			case Timer_stop:
				timer_pointer_states.Stop()
			case Timer_reset:
			timer_pointer_states.Reset(time.Duration(STATES_TIMEOUT_DURATION_MS)*(time.Millisecond))
			}

		case <- timer_pointer_states.C:
			timer_states_timeout <- true

		default:
		}
	}
}

func Timer_reAlivePeer_CabAgreement(timer_reAlivePeer_CabAgreement chan Timer_enum, timer_reAlivePeer_CabAgreement_timeout chan bool){
	timer_pointer_reAlivePeer_CabAgreement := time.NewTimer(time.Duration(REALIVE_PEER_CABAGREEMENT_TIMEOUT_MS)*(time.Millisecond))
	timer_pointer_reAlivePeer_CabAgreement.Stop()

	for{
		select{
		case timer_info := <- timer_reAlivePeer_CabAgreement:
			switch(timer_info){
			case Timer_stop:
				timer_pointer_reAlivePeer_CabAgreement.Stop()
			case Timer_reset:
			timer_pointer_reAlivePeer_CabAgreement.Reset(time.Duration(REALIVE_PEER_CABAGREEMENT_TIMEOUT_MS)*(time.Millisecond))
			}

		case <- timer_pointer_reAlivePeer_CabAgreement.C:
			timer_reAlivePeer_CabAgreement_timeout <- true

		default:
		}
	}
}

func Timer_detectImmobility(timer_detectImmobility chan Timer_enum, timer_detectImmobility_timeout chan bool){
	timer_pointer_detectImmobility := time.NewTimer(time.Duration(DETECT_IMMOBILITY_TIMEOUT_DURATION_MS)*(time.Millisecond))
	timer_pointer_detectImmobility.Stop()

	for{
		select{
		case timer_info := <- timer_detectImmobility:
			switch(timer_info){
			case Timer_stop:
				timer_pointer_detectImmobility.Stop()
			case Timer_reset:
			timer_pointer_detectImmobility.Reset(time.Duration(DETECT_IMMOBILITY_TIMEOUT_DURATION_MS)*(time.Millisecond))
			}

		case <- timer_pointer_detectImmobility.C:
			timer_detectImmobility_timeout <- true

		default:
		}
	}
}

