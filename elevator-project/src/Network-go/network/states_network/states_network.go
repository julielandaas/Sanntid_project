package states_network

import (
	"Sanntid/Network-go/network/conn"
	"Sanntid/Network-go/network/bcast"
	"fmt"
	"net"
	"sort"
	"time"
)

func send_states(){
	for{
		bcast.Transmitter()
	}
}