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
			fmt.Printf("timer_pointer_door.C: %+v\n", timeout)
			timer_open_door_timeout <- true

		default:
			//nothing happens
		}

	}
}