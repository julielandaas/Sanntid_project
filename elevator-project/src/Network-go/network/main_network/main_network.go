package main_network

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
	"Sanntid/Network-go/network/bcast"
	"Sanntid/Network-go/network/localip"
	"Sanntid/Network-go/network/peers"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

func states_MapToString(states map[string]requests.HRAElevState) string{
	jsonBytes, err := json.Marshal(states)
    	if err != nil {
        	fmt.Println("error in states_MapToString: json.Marshal error: ", err)
        	return ""
    	}
	
	return string(jsonBytes)
}

func states_StringToMap(message string) map[string]requests.HRAElevState{
	recieved := new(map[string]requests.HRAElevState)

	//KANSKJE []byte(message) er feil
	err := json.Unmarshal([]byte(message), &recieved)
    	if err != nil {
        	fmt.Println("json.Unmarshal error: ", err)
        	return nil
    	}

	return *recieved
}

func Main_network(requests_state_network chan elevio.Elevator, input_buttons_network chan elevio.ButtonEvent) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	hallRequests := [][2]bool{{false, false}, {false, false}, {false, false}, {false, false}}
	var newHallRequest elevio.ButtonEvent

	states := map[string]requests.HRAElevState{
		id: requests.HRAElevState{
			Behaviour:       "idle",
			Floor:          0,
			Direction:      "stop",
			CabRequests:    []bool{false, false, false, false},
		}}


	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	//peerUpdate_Hallreqs := make(chan peers.PeerUpdate)
	//peerUpdate_States := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)

	statesTx := make(chan HelloMsg)
	statesRx := make(chan HelloMsg)


	hallrequestsTx := make(chan HelloMsg)
	hallrequestsRx := make(chan HelloMsg)
	
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, statesTx)
	go bcast.Receiver(16569, statesRx)

	go bcast.Transmitter(20014, hallrequestsTx)
	go bcast.Receiver(20014, hallrequestsRx)


	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case button_request := <- input_buttons_network:
			if button_request.Button == elevio.BT_Cab{
				myState := states[id]
				myState.CabRequests[button_request.Floor] = true
				states[id] = myState
			}else{
				hallRequests[button_request.Floor][button_request.Button] = true
				//sende en update til netttverket
				newHallRequest = button_request
				//gjøre om til string typ json og sende
			}
			
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
		

		case mystate := <- requests_state_network:
			var cabRequests_temp []bool

			for f := 0; f < elevio.N_FLOORS; f++ {
				if mystate.Requests[f][elevio.BT_Cab] == true{
					cabRequests_temp[f] = true
					//input.States["one"].CabRequests[myState.Floor] = true
				}
			}

			states[id] = requests.HRAElevState{
				Behaviour:      elevio.Elevio_behaviour_toString(mystate.Behaviour),
				Floor:          mystate.Floor,
				Direction:      elevio.Elevio_dirn_toString(mystate.Dirn),
				CabRequests:    cabRequests_temp,
			}

			statesMsg := HelloMsg{states_MapToString(states), 0};
			statesMsg.Iter++
			statesTx <- statesMsg
		
		case msg := <- statesRx:
			states_StringToMap(msg.Message)

			/*



			*/
		}

		
	}
}

