package main_network

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
	"Sanntid/Driver-go/timer"
	"Sanntid/Network-go/network/bcast"
	"Sanntid/Network-go/network/localip"
	"Sanntid/Network-go/network/peers"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	//"timer"
	"reflect"
	//"strings"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//
//	will be received as zero-values.
type HelloMsg struct {
	Id      string
	Message string
	Iter    int
}

type IdIterpair struct {
	Id      string
	Message string
	Iter    int
}

func states_MapToString(states map[string]requests.HRAElevState) string {
	jsonBytes, err := json.Marshal(states)
	if err != nil {
		fmt.Println("error in states_MapToString: json.Marshal error: ", err)
		return ""
	}

	return string(jsonBytes)
}

func states_StringToMap(message string) map[string]requests.HRAElevState {
	recieved := new(map[string]requests.HRAElevState)

	//KANSKJE []byte(message) er feil
	err := json.Unmarshal([]byte(message), &recieved)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return *recieved
}

func buttonEvent_StructToString(buttonEvent elevio.ButtonEvent) string {
	jsonBytes, err := json.Marshal(buttonEvent)
	if err != nil {
		fmt.Println("error in states_MapToString: json.Marshal error: ", err)
		return ""
	}

	return string(jsonBytes)
}

func buttonEvent_StringToStruct(message string) *elevio.ButtonEvent {
	recieved := new(elevio.ButtonEvent)

	//KANSKJE []byte(message) er feil
	err := json.Unmarshal([]byte(message), &recieved)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return recieved
}

func Main_network(requests_state_network chan elevio.Elevator, input_buttons_network chan elevio.ButtonEvent, network_hallrequest_requests chan elevio.ButtonEvent,
	network_statesMap_requests chan map[string]requests.HRAElevState, network_id_requests chan string, requests_deleteHallRequest_network chan elevio.ButtonEvent, 
	timer_requests chan timer.Timer_enum,timer_requests_timeout chan bool, requests_timeout_duration_s int, timer_delete chan timer.Timer_enum,timer_delete_timeout chan bool, delete_timeout_duration_s int) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	var peersList []string
	
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
	network_id_requests <- id


	//var unconfirmed_hallRequests map[elevio.ButtonEvent][]string
	unconfirmed_hallRequests := make(map[elevio.ButtonEvent][]string)
	unconfirmed_hallDeletes := make(map[elevio.ButtonEvent][]string)
	var stateAgreementLst []string

	//unconfirmed_hallRequestsLst := [][2]bool{{false, false}, {false, false}, {false, false}, {false, false}}

	//var newHallRequest elevio.ButtonEvent
	states := make(map[string]requests.HRAElevState)
	states[id] = requests.HRAElevState{
			Behaviour:   "idle",
			Floor:       0,
			Direction:   "stop",
			CabRequests: [elevio.N_FLOORS]bool{false, false, false, false},
		}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	//peerUpdate_Hallreqs := make(chan peers.PeerUpdate)
	//peerUpdate_States := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15747, id, peerTxEnable)
	go peers.Receiver(15747, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	//helloTx := make(chan HelloMsg)
	//helloRx := make(chan HelloMsg)

	statesTx := make(chan HelloMsg)
	statesRx := make(chan HelloMsg)

	hallRequestsTx := make(chan HelloMsg)
	hallRequestsRx := make(chan HelloMsg)

	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, statesTx)
	go bcast.Receiver(16569, statesRx)

	go bcast.Transmitter(20014, hallRequestsTx)
	go bcast.Receiver(20014, hallRequestsRx)

	// The example message. We just send one of these every second.
	/*
	go func() {
		helloMsg := HelloMsg{id, "Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()
	*/
	fmt.Println("Started")
	for {
		select {
		case button_request := <-input_buttons_network:

			if button_request.Button == elevio.BT_Cab {
				myState := states[id]
				myState.CabRequests[button_request.Floor] = true
				states[id] = myState

				stateAgreementLst = nil
				stateAgreementLst = append(stateAgreementLst, id)

				statesMsg := HelloMsg{id, states_MapToString(states), 0}
				statesMsg.Iter++
				statesTx <- statesMsg

			} else {
				//unconfirmed_hallRequestsLst[button_request.Floor][button_request.Button] = true
				_, ok := unconfirmed_hallRequests[button_request]
				if !ok {
					unconfirmed_hallRequests[button_request]  = append(unconfirmed_hallRequests[button_request],id)
				}
				

				//fmt.Println("Recieved button press in network module")
				//sende en update til netttverket
				//newHallRequest = button_request
				hallRequestMsg := HelloMsg{id, buttonEvent_StructToString(button_request), 0}
				hallRequestMsg.Iter++
				hallRequestsTx <- hallRequestMsg

				timer_requests <- timer.Timer_stop
				timer_requests <- timer.Timer_reset

			}
		

		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			peersList = p.Peers
/*
			peer_id = nil
			for i := 0; i < len(p.Peers); i++ {
				temp_peer_id :=  (p.Peers[i])[strings.LastIndex((p.Peers[i]), "-") + 1:]
				peer_id = append(peer_id,temp_peer_id)			
			}
			fmt.Printf("  Peer_id:    %q\n", peer_id)
			*/
		/*
		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
		*/

		case mystate := <-requests_state_network:
			var cabRequests_temp [elevio.N_FLOORS]bool

			for f := 0; f < elevio.N_FLOORS; f++ {
				if mystate.Requests[f][elevio.BT_Cab] == true {
					cabRequests_temp[f] = true
					//input.States["one"].CabRequests[myState.Floor] = true
				}
			}

			states[id] = requests.HRAElevState{
				Behaviour:   elevio.Elevio_behaviour_toString(mystate.Behaviour),
				Floor:       mystate.Floor,
				Direction:   elevio.Elevio_dirn_toString(mystate.Dirn),
				CabRequests: cabRequests_temp,
			}
			stateAgreementLst = nil
			stateAgreementLst = append(stateAgreementLst,id)

			statesMsg := HelloMsg{id, states_MapToString(states), 0}
			statesMsg.Iter++
			statesTx <- statesMsg

		case statesMsg := <-statesRx:
			statesRecieved := states_StringToMap(statesMsg.Message)

			_, ok := states[statesMsg.Id]
			if !ok {
				states[statesMsg.Id] = statesRecieved[statesMsg.Id]
				fmt.Printf("new state sent back\n")
				statesMsg := HelloMsg{id, states_MapToString(states), 0}
				statesMsg.Iter++
				statesTx <- statesMsg
			}
			
			if statesMsg.Id != id {
				if !reflect.DeepEqual(states[statesMsg.Id], statesRecieved[statesMsg.Id]){
					fmt.Printf("new state has changed\n")
					states[statesMsg.Id] = statesRecieved[statesMsg.Id]
					
					statesMsg := HelloMsg{id, states_MapToString(states), 0}
					statesMsg.Iter++
					statesTx <- statesMsg
				}

				states[statesMsg.Id] = statesRecieved[statesMsg.Id]
				fmt.Printf("recieved state from %+v\n", statesMsg.Id)
				fmt.Printf("%+v\n", statesMsg.Message)
			
				if reflect.DeepEqual(states[id], statesRecieved[id]){
					flag:= false
					fmt.Printf("states are deep equal\n")
					for i := 0; i < len(stateAgreementLst); i++ {
						if stateAgreementLst[i] == statesMsg.Id{
							flag = true
						}
					}
					if !flag{
						stateAgreementLst = append(stateAgreementLst, statesMsg.Id)
					}
					//fmt.Printf("stateAgreemenLst length: %+v\n", len(stateAgreementLst))
				
				}
			}

			if len(stateAgreementLst) == len(peersList){
				println("Agreed on states\n")
				//sende til request assigner, KANSKJE FLYTTE
				network_statesMap_requests <- states
			} 
			//lagre de andre sine states
			//og oppdatere hvis vi ikke har de
			//og sjekke om alle har fått cabrequestsene våre
			//hvis vi ikke haar en state kan vi hente den fra nettverket
			//sende states til requestsassigner

			//fmt.Println("Recieved state from peer----------------")
		case deleted_Hallrequest := <- requests_deleteHallRequest_network:
			//fmt.Printf("\ntrying to delete hall\n")
			

			_, ok := unconfirmed_hallDeletes[deleted_Hallrequest]
				if !ok {
					unconfirmed_hallDeletes[deleted_Hallrequest]  = append(unconfirmed_hallRequests[deleted_Hallrequest],id)
				}
				

				//fmt.Println("Recieved button press in network module")
				//sende en update til netttverket
				//newHallRequest = button_request
				hallDeleteMsg := HelloMsg{id, buttonEvent_StructToString(deleted_Hallrequest), 0}
				hallDeleteMsg.Iter++
				hallRequestsTx <- hallDeleteMsg

				timer_delete <- timer.Timer_stop
				timer_delete <- timer.Timer_reset


		case hallRequest := <-hallRequestsRx:
			//fmt.Printf(hallRequest.Id)
			hallreq := *buttonEvent_StringToStruct(hallRequest.Message)
			fmt.Printf("Recieved hall request from peer %+v, floor: %+v\n",hallRequest.Id, hallreq.Floor)
			if hallRequest.Id != id {
				switch hallreq.Toggle{
				case true:
					_, ok := unconfirmed_hallRequests[hallreq]
					if ok {
						fmt.Printf("confirming recieved hall request\n")
						flag := 0
						for i := 0; i < len(unconfirmed_hallRequests[hallreq]); i++ {
							if hallRequest.Id == unconfirmed_hallRequests[hallreq][i]{
								flag = 1
							}
						} 
						if flag != 1{
							unconfirmed_hallRequests[hallreq] = append(unconfirmed_hallRequests[hallreq], hallRequest.Id)
						}
					}else{
						fmt.Printf("sending request back\n")
						network_hallrequest_requests <- hallreq

						hallRequest.Id = id
						hallRequestsTx <- hallRequest
					} 

				case false:
					_, ok := unconfirmed_hallDeletes[hallreq]
					if ok {
						flag := 0
						for i := 0; i < len(unconfirmed_hallDeletes[hallreq]); i++ {
							if hallRequest.Id == unconfirmed_hallDeletes[hallreq][i]{
								flag = 1
							}
						} 
						if flag != 1{
							unconfirmed_hallDeletes[hallreq] = append(unconfirmed_hallDeletes[hallreq], hallRequest.Id)
						}
					}else{
						network_hallrequest_requests <- hallreq

						hallRequest.Id = id
						hallRequestsTx <- hallRequest
					} 
				}
			}

			if len(unconfirmed_hallRequests[hallreq]) == len(peersList){
				fmt.Printf("everyone has recieved the order")
				network_hallrequest_requests <- hallreq
				delete(unconfirmed_hallRequests, hallreq)
				//stoppe timer
				//unconfirmed_hallRequestsLst[hallreq.Floor][hallreq.Button] = false
				
			}

			if len(unconfirmed_hallDeletes[hallreq]) == len(peersList){
				fmt.Printf("everyone has recieved the delete")
				//network_hallrequest_requests <- hallreq
				delete(unconfirmed_hallDeletes, hallreq)
				//stoppe timer
				//unconfirmed_hallRequestsLst[hallreq.Floor][hallreq.Button] = false
				
			}
		
		
		case <- timer_requests_timeout:
			//sjekk om unconfirmed hallrequests er empty,   hvis den ikke er det må vi gå gjennom requestsene i unconfirmed requests og sende de på nytt, deretter starte timeren igjen
			
			if len(unconfirmed_hallRequests) != 0 {
				for key, _ := range unconfirmed_hallRequests{
					hallRequestMsg := HelloMsg{id, buttonEvent_StructToString(key), 0}
					hallRequestMsg.Iter++
					hallRequestsTx <- hallRequestMsg
				}

				timer_requests <- timer.Timer_stop
				timer_requests <- timer.Timer_reset

			}
		case <- timer_delete_timeout:
			//sjekk om unconfirmed hallrequests er empty,   hvis den ikke er det må vi gå gjennom requestsene i unconfirmed requests og sende de på nytt, deretter starte timeren igjen
			
			if len(unconfirmed_hallDeletes) != 0 {
				for key, _ := range unconfirmed_hallDeletes{
				hallDeleteMsg := HelloMsg{id, buttonEvent_StructToString(key), 0}
				hallDeleteMsg.Iter++
				hallRequestsTx <- hallDeleteMsg
				}

				timer_delete <- timer.Timer_stop
				timer_delete <- timer.Timer_reset

			}
		


		default:

		}
	}
}