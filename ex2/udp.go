package main

import(
	"fmt"
	"net"
	"time"
)


func udp_recieve() {
	// Define the server address and port
	addr := net.UDPAddr{
		Port: 20014,
		IP:   net.ParseIP("10.100.23.24"),
	}

	// Create a UDP socket
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer conn.Close()
	fmt.Printf("Listening on %s\n", conn.LocalAddr().String())

	// Buffer to store received data
	buffer := make([]byte, 1024)

	for {
		// Receive data
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving:", err.Error())
			continue
		}
		fmt.Printf("Received %d bytes from %s: %s\n", n, remoteAddr, string(buffer[:n]))

		

		// Process data (for now, just print it)
		// Add your processing code here
	}
}

//With assistance from chatGPT

/*
Import Necessary Packages: Primarily, you'll need to import the net package.

Listen on a UDP Port: Use the net.ListenUDP function to create a UDP socket and listen for incoming datagrams.

Receive Data: Use the ReadFromUDP method to receive data from the network.

Process the Data: Once you've received the data, you can process it as needed.

Close the Connection: After you're done, close the UDP connection.
*/

func udp_send(){
	addr := net.UDPAddr{
		Port: 20014,
		IP:   net.ParseIP("10.100.23.255"),
	}

	// Create a UDP socket
	
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	defer conn.Close()
	fmt.Printf("Dialing %s\n", conn.LocalAddr().String())

	// Buffer to store received data
	//buffer := make([]byte, 0)
	//buffer = append(buffer, 7)
	for {
		// Receive data

		
		n, err := conn.Write([]byte("Hello from 14"))
		if err != nil {
			fmt.Println("Error sending:", err.Error())
			continue
		}
		fmt.Printf("sent %d bytes\n",n)

		time.Sleep(2*time.Second)

		// Process data (for now, just print it)
		// Add your processing code here
	}
}

func main(){

	go udp_recieve()
	
	go udp_send()
	for {

	}
	//time.Sleep(5*time.Second)
}