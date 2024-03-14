package main_network

import (
	"Sanntid/Driver-go/elevio"
	"Sanntid/Driver-go/requests"
	"encoding/json"
	"fmt"

)


func statesMap_MapToString(statesMap map[string]requests.HRAElevState) string {
	jsonBytes, err := json.Marshal(statesMap)
	if err != nil {
		fmt.Println("error in states_MapToString: json.Marshal error: ", err)
		return ""
	}

	return string(jsonBytes)
}

func statesMap_StringToMap(message string) map[string]requests.HRAElevState {
	stateMap := new(map[string]requests.HRAElevState)

	err := json.Unmarshal([]byte(message), &stateMap)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return *stateMap
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
	buttonEvent := new(elevio.ButtonEvent)

	err := json.Unmarshal([]byte(message), &buttonEvent)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return buttonEvent
}


func delete_recievedFromId_reInitCab_AckLst(recieved_Id string, reInitCab_AckLst []string) []string{
	var temp_reInitCab_AckLst []string
		for i := 0; i < len(reInitCab_AckLst); i++ {
			if reInitCab_AckLst[i] != recieved_Id  {
				temp_reInitCab_AckLst = append(temp_reInitCab_AckLst, reInitCab_AckLst[i])
			}
		}
	return temp_reInitCab_AckLst
}


func updateStatesMap(id string, mystate elevio.Elevator, statesMap map[string]requests.HRAElevState) {
	cabRequests_temp := [elevio.N_FLOORS]bool{false, false, false, false}

		for f := 0; f < elevio.N_FLOORS; f++ {
			if mystate.Requests[f][elevio.BT_Cab] {
				cabRequests_temp[f] = true
			}
		}

		statesMap[id] = requests.HRAElevState{
			Behaviour:   elevio.Elevio_behaviour_toString(mystate.Behaviour),
			Floor:       mystate.Floor,
			Direction:   elevio.Elevio_dirn_toString(mystate.Dirn),
			CabRequests: cabRequests_temp,
		}
}
