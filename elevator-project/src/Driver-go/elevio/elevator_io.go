package elevio

import (
	"Sanntid/Network-go/network/localip"
	"time"
	"sync"
	"net"
	"fmt"
	"os"
)


const POLLRATE = 20 * time.Millisecond

var initialized    bool = false
var numFloors      int = N_FLOORS
var mtx            sync.Mutex
var conn           net.Conn


func Init(addr string, id string) string {
	if initialized {
		fmt.Println("Driver already initialized!")
		return id
	} else {
		mtx = sync.Mutex{}
		var err error
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			panic(err.Error())
		}
		initialized = true

		if id == "" {
			localIP, err := localip.LocalIP()
			if err != nil {
				fmt.Println(err)
				localIP = "DISCONNECTED"
			}
			id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
		}
		return id
	}
}


func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}


func PollButtons(drv_buttonEvent_input chan<- ButtonEvent) {
	prev := make([][3]bool, numFloors)
	for {
		time.Sleep(POLLRATE)
		for f := 0; f < numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v != false {
					drv_buttonEvent_input <- ButtonEvent{f, ButtonType(b), true}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(drv_floor_input chan<- int) {
	prev := -1
	for {
		time.Sleep(POLLRATE)
		v := GetFloor()
		if v != prev && v != -1 {
			drv_floor_input <- v
		}
		prev = v
	}
}

func PollStopButton(drv_stop_input chan<- bool) {
	prev := false
	for {
		time.Sleep(POLLRATE)
		v := GetStop()
		if v != prev {
			drv_stop_input <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(drv_obstruction_input chan<- bool) {
	prev := false
	for {
		time.Sleep(POLLRATE)
		v := GetObstruction()
		if v != prev {
			drv_obstruction_input <- v
		}
		prev = v
	}
}

func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}


func read(in [4]byte) [4]byte {
	mtx.Lock()
	defer mtx.Unlock()
	
	_, err := conn.Write(in[:])
	if err != nil { panic("Lost connection to Elevator Server") }
	
	var out [4]byte
	_, err = conn.Read(out[:])
	if err != nil { panic("Lost connection to Elevator Server") }
	
	return out
}

func write(in [4]byte) {
	mtx.Lock()
	defer mtx.Unlock()
	
	_, err := conn.Write(in[:])
	if err != nil { panic("Lost connection to Elevator Server") }
}


func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}


