package timer

import(
	"time"
	"fmt"
)


type Timer_enum int
const (
	Timer_stop     Timer_enum = 1
	Timer_reset               = -1  
)


//enndre navnet på denne til noe med open door
func Timer_handler(timer_open_door chan Timer_enum,timer_open_door_timeout chan bool, doorOpenDuration_s int){
	timer_pointer_door := time.NewTimer(time.Duration(doorOpenDuration_s)*(time.Second))
	timer_pointer_door.Stop()

	for{
		select{
		case timer_info := <- timer_open_door:
			switch(timer_info){
			case Timer_stop:
				timer_pointer_door.Stop()
			case Timer_reset:
			timer_pointer_door.Reset(time.Duration(doorOpenDuration_s)*(time.Second))
			}

		case timeout := <- timer_pointer_door.C:
			fmt.Printf("timeout door: %+v\n", timeout)
			timer_open_door_timeout <- true

		default:
			//nothing happens
		}

	}
}

func Timer_Requests(timer_requests chan Timer_enum,timer_requests_timeout chan bool, requests_timeout_duration_s int){
	timer_pointer_requests := time.NewTimer(time.Duration(requests_timeout_duration_s)*(time.Second))
	timer_pointer_requests.Stop()

	for{
		select{
		case timer_info := <- timer_requests:
			switch(timer_info){
			case Timer_stop:
				timer_pointer_requests.Stop()
			case Timer_reset:
			timer_pointer_requests.Reset(time.Duration(requests_timeout_duration_s)*(time.Second))
			}

		case timeout := <- timer_pointer_requests.C:
			fmt.Printf("timerout requests %+v\n", timeout)
			timer_requests_timeout <- true

		default:
			//nothing happens
		}

	}
}

func Timer_deleteRequests(timer_delete chan Timer_enum,timer_delete_timeout chan bool, delete_timeout_duration_s int){
	timer_pointer_requests := time.NewTimer(time.Duration(delete_timeout_duration_s)*(time.Second))
	timer_pointer_requests.Stop()

	for{
		select{
		case timer_info := <- timer_delete:
			switch(timer_info){
			case Timer_stop:
				timer_pointer_requests.Stop()
			case Timer_reset:
			timer_pointer_requests.Reset(time.Duration(delete_timeout_duration_s)*(time.Second))
			}

		case timeout := <- timer_pointer_requests.C:
			fmt.Printf("timeout delete: %+v\n", timeout)
			timer_delete_timeout <- true

		default:
			//nothing happens
		}

	}
}
