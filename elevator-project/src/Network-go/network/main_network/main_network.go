package main_network

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
	"Sanntid/Driver-go/timer"
	"Sanntid/Network-go/network/bcast"
	"Sanntid/Network-go/network/localip"
	"Sanntid/Network-go/network/peers"
	"encoding/json"
	//"flag"
	"fmt"
	"os"
	//"time"
	"reflect"
	//"strings"
	"sync"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//

var statesMutex sync.Mutex
var deadPeerMapMutex sync.Mutex

// will be received as zero-values.
type HelloMsg struct {
	Id      string
	Message string
	Iter    int
}

type MessageID int
const (
	   ButtonType = 0
	BT_HallDown             = 1
	BT_Cab                  = 2
)


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
	timer_requests chan timer.Timer_enum, timer_requests_timeout chan bool, timer_delete chan timer.Timer_enum, timer_delete_timeout chan bool,
	timer_states chan timer.Timer_enum, timer_states_timeout chan bool, id string, network_peersList_requests chan []string,
	requests_resendHallrequests_network chan elevio.ButtonEvent, timer_reAlivePeer_CabAgreement chan timer.Timer_enum, timer_reAlivePeer_CabAgreement_timeout chan bool) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	peersList := []string {}
	RetrievedCabrequests_flag := false // skal motta cabreq fra andre heiser ved oppstart, viss man har dødd
	

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
	var stateAgreementLst []string                         // inneholder id-ene til de heisene som er enig med staten vår
	deadPeersMap := make(map[string][elevio.N_FLOORS]bool) // mapper cabreqs på id-ene
	
	reAlivePeer_CabAgreementLst := []string {} // inneholder id-ene til heisene som kommer på nettet igjen etter å ha vært døde

	//unconfirmed_hallRequestsLst := [][2]bool{{false, false}, {false, false}, {false, false}, {false, false}}

	//var newHallRequest elevio.ButtonEvent
	states := make(map[string]requests.HRAElevState)
	states[id] = requests.HRAElevState{
		Behaviour:   "idle",
		Floor:       0,
		Direction:   "stop",
		CabRequests: [elevio.N_FLOORS]bool{false, false, false, false},
	}

	//time.Sleep(500*time.Millisecond)

	// We make a channel for receiving updates on the id's of the peers that are alive on the network
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
	//fmt.Println("Started")
	for {
		select {
		case button_request := <-input_buttons_network:
			//fmt.Printf("recieved on input_button_network\n")

			if button_request.Button == elevio.BT_Cab {
				state_temp := states[id]
				state_temp.CabRequests[button_request.Floor] = true
				states[id] = state_temp

				if len(peersList) == 0{
					statesMutex.Lock()
					stateCopy := make(map[string]requests.HRAElevState)
					for k, v := range states {
						stateCopy[k] = v
					}
					statesMutex.Unlock()
					network_statesMap_requests <- stateCopy

					timer_requests <- timer.Timer_stop
					timer_delete <- timer.Timer_stop
					timer_states <- timer.Timer_stop

				}else{
					stateAgreementLst = nil
					stateAgreementLst = append(stateAgreementLst, id)

					statesMsg := HelloMsg{id, states_MapToString(states), 0}
					statesMsg.Iter++
					statesTx <- statesMsg

					

					timer_states <- timer.Timer_stop
					timer_states <- timer.Timer_reset
					//fmt.Printf("state timer started\n")
					//fmt.Printf("in button_request: %s\n", states_MapToString(states))				
				}

			} else {
				//unconfirmed_hallRequestsLst[button_request.Floor][button_request.Button] = true
				if len(peersList) > 0 {
					_, ok := unconfirmed_hallRequests[button_request]
					if !ok {
						unconfirmed_hallRequests[button_request] = append(unconfirmed_hallRequests[button_request], id)

						hallRequestMsg := HelloMsg{id, buttonEvent_StructToString(button_request), 0}
						hallRequestMsg.Iter++
						hallRequestsTx <- hallRequestMsg

						timer_requests <- timer.Timer_stop
						timer_requests <- timer.Timer_reset
					}
				}

				//fmt.Println("Recieved button press in network module")
				//sende en update til netttverket
				//newHallRequest = button_requesT

			}
		case pseudo_buttonevent := <- requests_resendHallrequests_network:
			_, ok := unconfirmed_hallRequests[pseudo_buttonevent]
				if !ok {
					unconfirmed_hallRequests[pseudo_buttonevent] = append(unconfirmed_hallRequests[pseudo_buttonevent], id)

					hallRequestMsg := HelloMsg{id, buttonEvent_StructToString(pseudo_buttonevent), 0}
					hallRequestMsg.Iter++
					hallRequestsTx <- hallRequestMsg

					timer_requests <- timer.Timer_stop
					timer_requests <- timer.Timer_reset
				}

		case p := <-peerUpdateCh:
			//fmt.Printf("recieved on peerupdatech\n")
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			peersList = p.Peers

			if len(p.Lost) > 0 {

				for i := 0; i < len(p.Lost); i++ {
					deadPeersMap[p.Lost[i]] = states[p.Lost[i]].CabRequests
					/*
						flag := false
						for j := 0; j < len(deadPeers); j++ {
							if p.Lost[i] == deadPeers[j]{
								flag = true
							}
						}
						if !flag {
							deadPeers = append(deadPeers, p.Lost[i])
						}*/
				}

			}

			network_peersList_requests <- peersList
			fmt.Printf("resending peersList: %+v\n", peersList)

			
			/*
			deadPeersMap_copy := make(map[string][elevio.N_FLOORS]bool)
			deadPeerMapMutex.Lock()
			for k, v := range deadPeersMap {
				deadPeersMap_copy[k] = v
			}
			deadPeerMapMutex.Unlock()
			*/
			


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
			//fmt.Printf("1. in mystate: %s\n", states_MapToString(states))

			cabRequests_temp := [elevio.N_FLOORS]bool{false, false, false, false}

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

			
			if len(peersList) == 0 {
				network_statesMap_requests <- states

				timer_requests <- timer.Timer_stop
				timer_delete <- timer.Timer_stop
				timer_states <- timer.Timer_stop
			} else {
				
				stateAgreementLst = nil
				stateAgreementLst = append(stateAgreementLst, id)
				//fmt.Printf("2. in mystate: %s\n", states_MapToString(states))
				statesMsg := HelloMsg{id, states_MapToString(states), 0}
				statesMsg.Iter++
				statesTx <- statesMsg

				timer_states <- timer.Timer_stop
				timer_states <- timer.Timer_reset
				//fmt.Printf("state timer started\n")
			}
			

		case statesMsg := <-statesRx:
			
			fmt.Printf("recieved on statesRx\n")
			statesMsg_recieved_id := statesMsg.Id
			statesMsg_recieved_iter := statesMsg.Iter
			statesMsg_recieved_states := states_StringToMap(statesMsg.Message)
			//fmt.Printf("1. in statesRx: %s\n", states_MapToString(states))
			/*
			for i := 0; i < len(reAlivePeer_CabAgreementLst); i++ {
				if reAlivePeer_CabAgreementLst[i] == statesMsg_recieved_id {
					if statesMsg_recieved_states[statesMsg_recieved_id].CabRequests == states[statesMsg_recieved_id].CabRequests ||
					   statesMsg_recieved_states[statesMsg_recieved_id].Floor {

					}
				}
			}*/
			if statesMsg_recieved_iter == 201{
				delete(deadPeersMap, statesMsg_recieved_id)
				var temp_reAlivePeer_CabAgreementLst []string
				for i := 0; i < len(reAlivePeer_CabAgreementLst); i++ {
					if reAlivePeer_CabAgreementLst[i] != statesMsg_recieved_id  {
						temp_reAlivePeer_CabAgreementLst = append(temp_reAlivePeer_CabAgreementLst, reAlivePeer_CabAgreementLst[i])
					}
				}
				reAlivePeer_CabAgreementLst = temp_reAlivePeer_CabAgreementLst
			}


			fmt.Printf("RetrievedCabrequests flag: %+v\n", RetrievedCabrequests_flag)
			fmt.Println("statesMsg_recieved_iter:", statesMsg_recieved_iter)


			if statesMsg_recieved_iter == 200 && !RetrievedCabrequests_flag {
				//states[statesMsg_recieved_id] = statesMsg_recieved_states[statesMsg_recieved_id]

				//fmt.Printf("Retrieved states: %+v\n", statesMsg_recieved_states)
				//time.Sleep(500*time.Millisecond)
				
				previous_Cabs := statesMsg_recieved_states[id].CabRequests
				//fmt.Printf("Retrieved Cabrequests: %+v, from %+v\n", previous_Cabs, statesMsg_recieved_id)

				temp_state := states[id]
				temp_state.CabRequests = previous_Cabs
				states[id] = temp_state

				for k, _ := range statesMsg_recieved_states {
					if (k != id) && (k != statesMsg_recieved_id) {
						flag := false
						for i := 0; i < len(peersList); i++ { // sjekk om tredjeparten er i live eller død
							if peersList[i] == k {
								flag = true
							}
						}

						if !flag { // viss denne er død, så stol på den som blir sendt til deg om han
							deadPeersMap[k] = statesMsg_recieved_states[k].CabRequests
							states[k] = statesMsg_recieved_states[k]
							fmt.Printf("State to dead peer: %+v", states_MapToString(states))
						}
					}

				}

				stateAgreementLst = nil
				stateAgreementLst = append(stateAgreementLst, id)

				timer_states <- timer.Timer_stop
				timer_states <- timer.Timer_reset

				statesMutex.Lock()
				stateCopy := make(map[string]requests.HRAElevState)
				for k, v := range states {
					stateCopy[k] = v
				}
				statesMutex.Unlock()

				network_statesMap_requests <- stateCopy
				
				//fmt.Printf("new state sent back, %+v\n", statesMsg.Message)

				statesMsg := HelloMsg{id, states_MapToString(states), 0}
				statesMsg.Iter = statesMsg_recieved_iter + 1
				statesTx <- statesMsg

				RetrievedCabrequests_flag = true

			} else if statesMsg_recieved_iter == 200{
				//her gjør den ingen av de andre tinegne da men kanskje det går greit
				statesMsg := HelloMsg{id, states_MapToString(states), 0}
				statesMsg.Iter = statesMsg_recieved_iter + 1
				statesTx <- statesMsg
			}
			

			_, ok := states[statesMsg_recieved_id]
			// mottar fra en heis vi ikkje har i map-et fra før av
			if !ok { 

				
				//fmt.Printf("2. in statesRx: %s\n", states_MapToString(states))

				// jeg er ny selv, og må få gamle cab-reqs tilbake.
				

				//_, myidInRecievedState := statesMsg_recieved_states[id]
				states[statesMsg_recieved_id] = statesMsg_recieved_states[statesMsg_recieved_id]

				statesMsg := HelloMsg{id, states_MapToString(states), 0}
				statesMsg.Iter++
				statesTx <- statesMsg


				
				//fmt.Printf("new state sent back, %+v\n", statesMsg.Message)

			// mottar fra noken andre, men den er IKKJE NY
			} else if statesMsg_recieved_id != id { 
				/*
					var deadPeers_temp []string
					for i := 0; i < len(deadPeers); i++ {
						if deadPeers[i] != statesMsg.Id {
							deadPeers_temp = append(deadPeers_temp, deadPeers[i])
						}
					}*/

				_, ok := deadPeersMap[statesMsg_recieved_id] // den har dødd, men vi har ikkje rukke å gjort noke, dermed heller ikkje sletta den fra map-et vårt
				if ok {
					
					for k, _ := range deadPeersMap {
						if (k != id) {
							temp_state := states[k]
							temp_state.CabRequests = deadPeersMap[k]
							states[k] = temp_state
						}
					}

					
					reAlivePeer_CabAgreementLst = append(reAlivePeer_CabAgreementLst, statesMsg_recieved_id)
					fmt.Printf("realivePeer_CabAgreement: %+v \n", reAlivePeer_CabAgreementLst)
					

					fmt.Printf("Peer alive, DeadPeers: %+v\n", deadPeersMap)
					fmt.Printf("Previous Cabrequests: %+v\n", states[statesMsg_recieved_id].CabRequests)

					statesMsg := HelloMsg{id, states_MapToString(states), 0}
					statesMsg.Iter = 200
					statesTx <- statesMsg

					timer_reAlivePeer_CabAgreement <- timer.Timer_stop
					timer_reAlivePeer_CabAgreement <- timer.Timer_reset

				
				} else if !reflect.DeepEqual(states[statesMsg_recieved_id], statesMsg_recieved_states[statesMsg_recieved_id]) { // dei gir oss ny informasjon om seg sjølv
					//fmt.Printf("new state has changed\n")
					//fmt.Printf("4. in statesRX: %s\n", states_MapToString(states))
					states[statesMsg_recieved_id] = statesMsg_recieved_states[statesMsg_recieved_id]
					//fmt.Printf("5. in statesRX: %s\n", states_MapToString(states))
					statesMsg := HelloMsg{id, states_MapToString(states), 0}
					statesMsg.Iter++
					statesTx <- statesMsg

				} else {
					if statesMsg_recieved_iter == 100 {
						statesMsg := HelloMsg{id, states_MapToString(states), 0}
						statesMsg.Iter++
						statesTx <- statesMsg
					}

				}

				//states[statesMsg.Id] = statesMsg_recieved_states[statesMsg.Id]
				//fmt.Printf("recieved state from %+v\n", statesMsg.Id)
				//fmt.Printf("%+v\n", statesMsg.Message)

				if reflect.DeepEqual(states[id], statesMsg_recieved_states[id]) {
					RetrievedCabrequests_flag = true
					flag := false
					//fmt.Printf("states are deep equal\n")
					for i := 0; i < len(stateAgreementLst); i++ {
						if stateAgreementLst[i] == statesMsg_recieved_id {
							flag = true
						}
					}
					if !flag {
						stateAgreementLst = append(stateAgreementLst, statesMsg_recieved_id)
					}
					//fmt.Printf("stateAgreemenLst length: %+v\n", len(stateAgreementLst))

				}
			}
			//fmt.Printf("6. in statesRX: %s\n", states_MapToString(states))
			///fmt.Printf("stateagreementlist: %+v\n", stateAgreementLst)
			//fmt.Printf("peerslist: %+v\n", peersList)


			if len(stateAgreementLst) == len(peersList) {
				//println("Agreed on states\n")
				//sende til request assigner, KANSKJE FLYTTE

				statesMutex.Lock()
				stateCopy := make(map[string]requests.HRAElevState)
				for k, v := range states {
					stateCopy[k] = v
				}
				statesMutex.Unlock()
				network_statesMap_requests <- stateCopy
				timer_states <- timer.Timer_stop
			}
			//lagre de andre sine states
			//og oppdatere hvis vi ikke har de
			//og sjekke om alle har fått cabrequestsene våre
			//hvis vi ikke haar en state kan vi hente den fra nettverket
			//sende states til requestsassigner

			//fmt.Println("Recieved state from peer----------------")
		case deleted_Hallrequest := <-requests_deleteHallRequest_network:
			//fmt.Printf("recieved on requsts_deletedhallrequest_network\n")
			//fmt.Printf("\ntrying to delete hall\n")

			if len(peersList) == 0{
				network_hallrequest_requests <- deleted_Hallrequest

				timer_requests <- timer.Timer_stop
				timer_delete <- timer.Timer_stop
				timer_states <- timer.Timer_stop
			}else{
				_, ok := unconfirmed_hallDeletes[deleted_Hallrequest]
				if !ok {
					unconfirmed_hallDeletes[deleted_Hallrequest] = append(unconfirmed_hallRequests[deleted_Hallrequest], id)
				}

				//fmt.Println("Recieved button press in network module")
				//sende en update til netttverket
				//newHallRequest = button_request
				hallDeleteMsg := HelloMsg{id, buttonEvent_StructToString(deleted_Hallrequest), 0}
				hallDeleteMsg.Iter++
				hallRequestsTx <- hallDeleteMsg

				timer_delete <- timer.Timer_stop
				timer_delete <- timer.Timer_reset
			}

			

		case hallRequest := <-hallRequestsRx:
			
			//fmt.Printf(hallRequest.Id)
			hallreq := *buttonEvent_StringToStruct(hallRequest.Message)
			//fmt.Printf("Recieved hall request from peer %+v, floor: %+v\n",hallRequest.Id, hallreq.Floor)
			if hallRequest.Id != id {
				switch hallreq.Toggle {
				case true:
					_, ok := unconfirmed_hallRequests[hallreq]
					if ok {
						//fmt.Printf("confirming recieved hall request\n")
						flag := 0
						for i := 0; i < len(unconfirmed_hallRequests[hallreq]); i++ {
							if hallRequest.Id == unconfirmed_hallRequests[hallreq][i] {
								flag = 1
							}
						}
						if flag != 1 {
							unconfirmed_hallRequests[hallreq] = append(unconfirmed_hallRequests[hallreq], hallRequest.Id)
						}
					} else {
						//fmt.Printf("sending request back\n")
						if hallRequest.Iter < 2 {
							network_hallrequest_requests <- hallreq

							hallRequest.Id = id
							hallRequest.Iter++
							hallRequestsTx <- hallRequest
						}

					}

				case false:
					_, ok := unconfirmed_hallDeletes[hallreq]
					if ok {
						flag := 0
						for i := 0; i < len(unconfirmed_hallDeletes[hallreq]); i++ {
							if hallRequest.Id == unconfirmed_hallDeletes[hallreq][i] {
								flag = 1
							}
						}
						if flag != 1 {
							unconfirmed_hallDeletes[hallreq] = append(unconfirmed_hallDeletes[hallreq], hallRequest.Id)
						}
					} else {
						if hallRequest.Iter < 2 {
							network_hallrequest_requests <- hallreq

							hallRequest.Id = id
							hallRequest.Iter++
							hallRequestsTx <- hallRequest
						}
					}
				}
			}
			
			if len(unconfirmed_hallRequests[hallreq]) == len(peersList) {
				//fmt.Printf("everyone has recieved the order")
				network_hallrequest_requests <- hallreq
				delete(unconfirmed_hallRequests, hallreq)
				//stoppe timer
				//timer_requests <- timer.Timer_stop
				//unconfirmed_hallRequestsLst[hallreq.Floor][hallreq.Button] = false

			}
			fmt.Printf("unconfirmedhallDeletes[hallreq] = %+v\n", unconfirmed_hallDeletes)
			fmt.Printf("peerslist = %+v\n", peersList)
			if len(unconfirmed_hallDeletes[hallreq]) == len(peersList) {
				//fmt.Printf("everyone has recieved the delete")
				//network_hallrequest_requests <- hallreq
				network_hallrequest_requests <- hallreq
				delete(unconfirmed_hallDeletes, hallreq)

				//timer_delete <- timer.Timer_stop

				//unconfirmed_hallRequestsLst[hallreq.Floor][hallreq.Button] = false

			}

		case <-timer_requests_timeout:
			
			//sjekk om unconfirmed hallrequests er empty,   hvis den ikke er det må vi gå gjennom requestsene i unconfirmed requests og sende de på nytt, deretter starte timeren igjen

			if len(unconfirmed_hallRequests) != 0 {
				for key, _ := range unconfirmed_hallRequests {
					hallRequestMsg := HelloMsg{id, buttonEvent_StructToString(key), 0}
					hallRequestMsg.Iter++
					hallRequestsTx <- hallRequestMsg
				}

				timer_requests <- timer.Timer_stop
				timer_requests <- timer.Timer_reset

			}else{
				timer_requests <- timer.Timer_stop
			}
		case <-timer_delete_timeout:
			
			//sjekk om unconfirmed hallrequests er empty,   hvis den ikke er det må vi gå gjennom requestsene i unconfirmed requests og sende de på nytt, deretter starte timeren igjen

			if len(unconfirmed_hallDeletes) != 0 {
				for key, _ := range unconfirmed_hallDeletes {
					hallDeleteMsg := HelloMsg{id, buttonEvent_StructToString(key), 0}
					hallDeleteMsg.Iter++
					hallRequestsTx <- hallDeleteMsg
				}

				timer_delete <- timer.Timer_stop
				timer_delete <- timer.Timer_reset

			}else{
				timer_delete <- timer.Timer_stop
			}
		case <-timer_states_timeout:
			
			//sjekk om unconfirmed hallrequests er empty,   hvis den ikke er det må vi gå gjennom requestsene i unconfirmed requests og sende de på nytt, deretter starte timeren igjen

			if (len(stateAgreementLst) < len(peersList)) || (len(peersList) == 1) {

				statesMsg := HelloMsg{id, states_MapToString(states), 0}
				statesMsg.Iter = 100
				statesTx <- statesMsg

				timer_states <- timer.Timer_stop
				timer_states <- timer.Timer_reset
			}
		case <- timer_reAlivePeer_CabAgreement_timeout: // USIKKER PÅ OM DETTE FUNKER
			if len(reAlivePeer_CabAgreementLst) > 0 {

				for i := 0; i < len(reAlivePeer_CabAgreementLst); i++ {
					temp_state := states[reAlivePeer_CabAgreementLst[i]]
					temp_state.CabRequests = deadPeersMap[reAlivePeer_CabAgreementLst[i]]
					states[reAlivePeer_CabAgreementLst[i]] = temp_state

					statesMsg := HelloMsg{id, states_MapToString(states), 0}
					statesMsg.Iter = 200
					statesTx <- statesMsg
					
					timer_reAlivePeer_CabAgreement <- timer.Timer_stop
					timer_reAlivePeer_CabAgreement <- timer.Timer_reset
				}
			} else {
				timer_reAlivePeer_CabAgreement <- timer.Timer_stop
			}

		default:

		}
	}
}
// må legge in at den nye heisen som gjenoppstår fra de døde må blokke alt av nye ordre fram til "fasiten" aka 200-meldinga er godkjent