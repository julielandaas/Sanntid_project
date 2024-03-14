package main_network

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
	"Sanntid/Driver-go/timer"
	"Sanntid/Network-go/network/bcast"
	"Sanntid/Network-go/network/peers"
	"fmt"
	"reflect"
	"sync"
)

var statesMutex sync.Mutex

type Message struct {
	Id            string
	Content       string
	Type          int
}

type MessageType int 
const (
	Normal 					 = 0
	Normal_Ack 				 = 1
	ReInitCab			     = 2
	ReInitCab_Ack            = 3
)


func Main_network(id string, fsm_state_network chan elevio.Elevator, input_buttons_network chan elevio.ButtonEvent, network_hallrequest_requests chan elevio.ButtonEvent,
	network_statesMap_requests chan map[string]requests.HRAElevState, fsm_deleteHallRequest_network chan elevio.ButtonEvent,
	timer_requests chan timer.Timer_enum, timer_requests_timeout chan bool, timer_delete chan timer.Timer_enum, timer_delete_timeout chan bool,
	timer_states chan timer.Timer_enum, timer_states_timeout chan bool, network_peersList_requests chan []string,
	requests_resendHallrequests_network chan elevio.ButtonEvent, timer_reInitCab_Ack chan timer.Timer_enum, timer_reInitCab_Ack_timeout chan bool) {


	statesMap := make(map[string]requests.HRAElevState)
	statesMap[id] = requests.HRAElevState{
		Behaviour:   "idle",
		Floor:       0,
		Direction:   "stop",
		CabRequests: [elevio.N_FLOORS]bool{false, false, false, false},
	}

	peersList := []string {}
	deadPeersMap := make(map[string][elevio.N_FLOORS]bool) 
	
	RetrievedMyCabrequests_flag := false 
	connectedToNetwork_flag := false

	unconfirmed_hallRequests := make(map[elevio.ButtonEvent][]string)
	unconfirmed_hallDeletes := make(map[elevio.ButtonEvent][]string)
	
	myState_AckLst := []string {}                     
	reInitCab_AckLst := []string {}


	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15747, id, peerTxEnable)
	go peers.Receiver(15747, peerUpdateCh)

	statesTx := make(chan Message)
	statesRx := make(chan Message)
	go bcast.Transmitter(16569, statesTx)
	go bcast.Receiver(16569, statesRx)

	hallRequestsTx := make(chan Message)
	hallRequestsRx := make(chan Message)
	go bcast.Transmitter(20014, hallRequestsTx)
	go bcast.Receiver(20014, hallRequestsRx)

	
	for {
		select {
		case button_request := <-input_buttons_network:
			if button_request.Button == elevio.BT_Cab {
				state_temp := statesMap[id]
				state_temp.CabRequests[button_request.Floor] = true
				statesMap[id] = state_temp

				if !connectedToNetwork_flag {
					statesMutex.Lock()
					stateCopy := make(map[string]requests.HRAElevState)
					for k, v := range statesMap {
						stateCopy[k] = v
					}
					statesMutex.Unlock()
					network_statesMap_requests <- stateCopy

					timer_requests <- timer.Timer_stop
					timer_delete <- timer.Timer_stop
					timer_states <- timer.Timer_stop

				}else {
					myState_AckLst = nil
					myState_AckLst = append(myState_AckLst, id)

					statesMsg := Message{id, statesMap_MapToString(statesMap), Normal}
					statesTx <- statesMsg					

					timer_states <- timer.Timer_stop
					timer_states <- timer.Timer_reset			
				}

			} else {
				if connectedToNetwork_flag {
					_, ok := unconfirmed_hallRequests[button_request]
					if !ok {
						unconfirmed_hallRequests[button_request] = append(unconfirmed_hallRequests[button_request], id)

						hallRequestMsg := Message{id, buttonEvent_StructToString(button_request), Normal}
						hallRequestsTx <- hallRequestMsg

						timer_requests <- timer.Timer_stop
						timer_requests <- timer.Timer_reset
					}
				}
			}

		case pseudo_HallButtonevent := <- requests_resendHallrequests_network:
			_, ok := unconfirmed_hallRequests[pseudo_HallButtonevent]
				if !ok {
					unconfirmed_hallRequests[pseudo_HallButtonevent] = append(unconfirmed_hallRequests[pseudo_HallButtonevent], id)

					hallRequestMsg := Message{id, buttonEvent_StructToString(pseudo_HallButtonevent), Normal}
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

			if len(peersList) == 0 {
				connectedToNetwork_flag = false
			} else {
				connectedToNetwork_flag = true
			}

			if len(p.Lost) > 0 {
				for i := 0; i < len(p.Lost); i++ {
					deadPeersMap[p.Lost[i]] = statesMap[p.Lost[i]].CabRequests
				}
			}
			network_peersList_requests <- peersList

		case mystate := <-fsm_state_network:
			updateStatesMap(id, mystate, statesMap)

			if !connectedToNetwork_flag {
				network_statesMap_requests <- statesMap

				timer_requests <- timer.Timer_stop
				timer_delete <- timer.Timer_stop
				timer_states <- timer.Timer_stop

			} else {
				myState_AckLst = nil
				myState_AckLst = append(myState_AckLst, id)

				statesMsg := Message{id, statesMap_MapToString(statesMap), Normal}
				statesTx <- statesMsg

				timer_states <- timer.Timer_stop
				timer_states <- timer.Timer_reset
			}

		case statesMsg := <-statesRx:
			recievedFrom_Id := statesMsg.Id
			recieved_Type := statesMsg.Type
			recieved_StatesMap := statesMap_StringToMap(statesMsg.Content)

			switch recieved_Type {
			case Normal:
				if RetrievedMyCabrequests_flag {
					_, ok := statesMap[recievedFrom_Id]
					if !ok { 
						statesMap[recievedFrom_Id] = recieved_StatesMap[recievedFrom_Id]
				
						statesMsg_Ack := Message{id, statesMap_MapToString(statesMap), Normal_Ack}
						statesTx <- statesMsg_Ack

						myState_AckLst = nil
						myState_AckLst = append(myState_AckLst, id)

						timer_states <- timer.Timer_stop
						timer_states <- timer.Timer_reset

						statesMsg := Message{id, statesMap_MapToString(statesMap), Normal}
						statesTx <- statesMsg

					}else if recievedFrom_Id != id {
						_, ok := deadPeersMap[recievedFrom_Id] 
						if ok {
							for key_id, _ := range deadPeersMap {
								if (key_id != id) {
									temp_state := statesMap[key_id]
									temp_state.CabRequests = deadPeersMap[key_id]
									statesMap[key_id] = temp_state
								}
							}
			
							reInitCab_AckLst = append(reInitCab_AckLst, recievedFrom_Id)

							statesMsg := Message{id, statesMap_MapToString(statesMap), ReInitCab}
							statesTx <- statesMsg
				
							timer_reInitCab_Ack <- timer.Timer_stop
							timer_reInitCab_Ack <- timer.Timer_reset
				
						
						}else {
							statesMap[recievedFrom_Id] = recieved_StatesMap[recievedFrom_Id]

							statesMsg := Message{id, statesMap_MapToString(statesMap), Normal_Ack}
							statesTx <- statesMsg
						}
					}else {
						if len(peersList) ==1{
							statesMsg := Message{id, statesMap_MapToString(statesMap), Normal_Ack}
							statesTx <- statesMsg
						}

					}
				} else {
					if recievedFrom_Id == id { 
						statesMap[recievedFrom_Id] = recieved_StatesMap[recievedFrom_Id]
				
						statesMsg_Ack := Message{id, statesMap_MapToString(statesMap), Normal_Ack}
						statesTx <- statesMsg_Ack
					}
				}

			case Normal_Ack:
				if (len(peersList) == 1 || (recievedFrom_Id != id && reflect.DeepEqual(statesMap[id], recieved_StatesMap[id]))) && !RetrievedMyCabrequests_flag {
					RetrievedMyCabrequests_flag = true
				}

				if RetrievedMyCabrequests_flag {
					if reflect.DeepEqual(statesMap[id], recieved_StatesMap[id]) {
						idIn_myState_AckLst_flag := false

	
						for i := 0; i < len(myState_AckLst); i++ {
							if myState_AckLst[i] == recievedFrom_Id {
								idIn_myState_AckLst_flag = true
							}
						}
	
						if !idIn_myState_AckLst_flag {
							myState_AckLst = append(myState_AckLst, recievedFrom_Id)
						}
					}
	
					if len(myState_AckLst) == len(peersList) {			
						statesMutex.Lock()
						stateCopy := make(map[string]requests.HRAElevState)
						for k, v := range statesMap {
							stateCopy[k] = v
						}
						statesMutex.Unlock()
						network_statesMap_requests <- stateCopy
						timer_states <- timer.Timer_stop
					}
				}
			
			case ReInitCab:
				if !RetrievedMyCabrequests_flag {
					RetrievedMyCabrequests_flag = true
					
					previous_Cabs := recieved_StatesMap[id].CabRequests
					temp_state := statesMap[id]
					temp_state.CabRequests = previous_Cabs
					statesMap[id] = temp_state
							
					// If there are other dead elevators, trust what the sender knows about the dead elevators
					for key_id, _ := range recieved_StatesMap {
						if (key_id != id) && (key_id != recievedFrom_Id) {
							peerAlive_flag := false
							for i := 0; i < len(peersList); i++ { 
								if peersList[i] == key_id {
									peerAlive_flag = true
								}
							}
							if !peerAlive_flag { 
								deadPeersMap[key_id] = recieved_StatesMap[key_id].CabRequests
								statesMap[key_id] = recieved_StatesMap[key_id]
							}
						}
					}
		
					myState_AckLst = nil
					myState_AckLst = append(myState_AckLst, id)
		
					timer_states <- timer.Timer_stop
					timer_states <- timer.Timer_reset
		
					statesMutex.Lock()
					stateCopy := make(map[string]requests.HRAElevState)

					for k, v := range statesMap {
						stateCopy[k] = v
					}
					statesMutex.Unlock()
					network_statesMap_requests <- stateCopy
				}
				statesMsg := Message{id, statesMap_MapToString(statesMap), ReInitCab_Ack}
				statesTx <- statesMsg


			case ReInitCab_Ack:
				delete(deadPeersMap, recievedFrom_Id)
				reInitCab_AckLst = delete_recievedFromId_reInitCab_AckLst(recievedFrom_Id, reInitCab_AckLst)
				if !reflect.DeepEqual(statesMap[id], recieved_StatesMap[id]){
					myState_AckLst = nil
					myState_AckLst = append(myState_AckLst, id)

					timer_states <- timer.Timer_stop
					timer_states <- timer.Timer_reset

					statesMsg := Message{id, statesMap_MapToString(statesMap), Normal}
					statesTx <- statesMsg
				}
			}
		
		case deleted_Hallrequest := <-fsm_deleteHallRequest_network:
			if !connectedToNetwork_flag {
				network_hallrequest_requests <- deleted_Hallrequest

				timer_requests <- timer.Timer_stop
				timer_delete <- timer.Timer_stop
				timer_states <- timer.Timer_stop

			}else {
				_, ok := unconfirmed_hallDeletes[deleted_Hallrequest]
				if !ok {
					unconfirmed_hallDeletes[deleted_Hallrequest] = append(unconfirmed_hallRequests[deleted_Hallrequest], id)
				}

				hallDeleteMsg := Message{id, buttonEvent_StructToString(deleted_Hallrequest), Normal}
				hallRequestsTx <- hallDeleteMsg

				timer_delete <- timer.Timer_stop
				timer_delete <- timer.Timer_reset
			}

		case hallRequestMsg := <-hallRequestsRx:
			recieved_hallrequest := *buttonEvent_StringToStruct(hallRequestMsg.Content)
			recieved_from_id  := hallRequestMsg.Id
			recieved_messageType := hallRequestMsg.Type
			
			switch recieved_hallrequest.Value {
			case true:
				if recieved_from_id != id {
					_, ok := unconfirmed_hallRequests[recieved_hallrequest]
					if ok {
						senderId_hasAckHallreq_prev := false
						for i := 0; i < len(unconfirmed_hallRequests[recieved_hallrequest]); i++ {
							if recieved_from_id == unconfirmed_hallRequests[recieved_hallrequest][i] {
								senderId_hasAckHallreq_prev = true
							}
						}
						if !senderId_hasAckHallreq_prev {
							unconfirmed_hallRequests[recieved_hallrequest] = append(unconfirmed_hallRequests[recieved_hallrequest], recieved_from_id)
						}
					} else {
						if recieved_messageType == Normal { 
							network_hallrequest_requests <- recieved_hallrequest
							hallRequestMsg_send := Message{id, buttonEvent_StructToString(recieved_hallrequest), Normal_Ack}
							hallRequestsTx <- hallRequestMsg_send
						}
					}
				}

				if len(unconfirmed_hallRequests[recieved_hallrequest]) == len(peersList) {
					network_hallrequest_requests <- recieved_hallrequest
					delete(unconfirmed_hallRequests, recieved_hallrequest)
				}

			case false:
				if recieved_from_id != id {
					_, ok := unconfirmed_hallDeletes[recieved_hallrequest]
					if ok {
						flag := 0
						for i := 0; i < len(unconfirmed_hallDeletes[recieved_hallrequest]); i++ {
							if recieved_from_id == unconfirmed_hallDeletes[recieved_hallrequest][i] {
								flag = 1
							}
						}
						if flag != 1 {
							unconfirmed_hallDeletes[recieved_hallrequest] = append(unconfirmed_hallDeletes[recieved_hallrequest], recieved_from_id)
						}
					} else {
						if recieved_messageType == Normal {
							network_hallrequest_requests <- recieved_hallrequest
							hallRequestMsg_send := Message{id, buttonEvent_StructToString(recieved_hallrequest), Normal_Ack}
							hallRequestsTx <- hallRequestMsg_send
						}
					}
				}
				
				if len(unconfirmed_hallDeletes[recieved_hallrequest]) == len(peersList) {
					network_hallrequest_requests <- recieved_hallrequest
					delete(unconfirmed_hallDeletes, recieved_hallrequest)
				}
			}

		case <-timer_requests_timeout:
			if len(unconfirmed_hallRequests) != 0 {
				for key_HallRequest, _ := range unconfirmed_hallRequests {
					hallRequestMsg := Message{id, buttonEvent_StructToString(key_HallRequest), Normal}
					hallRequestsTx <- hallRequestMsg
				}

				timer_requests <- timer.Timer_stop
				timer_requests <- timer.Timer_reset
			}

		case <-timer_delete_timeout:
			if len(unconfirmed_hallDeletes) != 0 {
				for key_HallRequest, _ := range unconfirmed_hallDeletes {
					hallDeleteMsg := Message{id, buttonEvent_StructToString(key_HallRequest), Normal}
					hallRequestsTx <- hallDeleteMsg
				}

				timer_delete <- timer.Timer_stop
				timer_delete <- timer.Timer_reset
			}

		case <-timer_states_timeout:
			if (len(myState_AckLst) < len(peersList)) || (len(peersList) == 1) {
				statesMsg := Message{id, statesMap_MapToString(statesMap), Normal}
				statesTx <- statesMsg

				timer_states <- timer.Timer_stop
				timer_states <- timer.Timer_reset
			}

		case <- timer_reInitCab_Ack_timeout: 
			if len(reInitCab_AckLst) > 0 {
				for i := 0; i < len(reInitCab_AckLst); i++ {
					temp_state := statesMap[reInitCab_AckLst[i]]
					temp_state.CabRequests = deadPeersMap[reInitCab_AckLst[i]]
					statesMap[reInitCab_AckLst[i]] = temp_state

					statesMsg := Message{id, statesMap_MapToString(statesMap), ReInitCab}
					statesTx <- statesMsg
					
					timer_reInitCab_Ack <- timer.Timer_stop
					timer_reInitCab_Ack <- timer.Timer_reset
				}
			}

		default:

		}
	}
}