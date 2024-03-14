 This program is made for running multiple elevators that communicates on the network.
 
 To run the program; 
 1. start an elevatorserver in a separate terminal,by using the command
 `elevatorserver`
 2. run this program by using the command
 `go run main.go -port "port" -id "our_id"`
 where 
  - "port" is the port of the elevatorserver, 15657
  - "our_id" is our chosen id for the elevator. All elevators on the network has to have an unique id.
  Both of these flags must be provided.