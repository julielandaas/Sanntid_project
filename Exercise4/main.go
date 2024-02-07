package main

import(
	"fmt"
	"os/exec"
	"net"
	"time"
	"strconv"
	"strings"
)

func dial_UDP() *net.UDPConn {
	addr := net.UDPAddr{
		Port: 20014,
		IP:   net.ParseIP("10.100.23.255"),
	}

	// Create a UDP socket
	
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return nil
	}
	//defer conn.Close()
	fmt.Printf("Dialing %s\n", conn.LocalAddr().String())

	return conn
}

func udp_send(count int, conn *net.UDPConn){
	

	// Buffer to store received data
	//buffer := make([]byte, 0)
	//buffer = append(buffer, 7)
	
	// Receive data

		
	n, err := conn.Write([]byte(fmt.Sprint(count)))
	if err != nil {
		fmt.Println("Error sending:", err.Error())
	}
	fmt.Printf("sent %d bytes\n",n)

	// Process data (for now, just print it)
	// Add your processing code here
	
}

func udp_recieve(errch chan int, recieved chan int) {
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

	err = conn.SetReadDeadline(time.Now().Add(1*time.Second))
	if err != nil {
		fmt.Println("Could not set deadline")
	}

	// Buffer to store received data
	buffer := make([]byte, 1024)

	for {
		// Receive data
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				fmt.Println("Timeout")
				errch <- 1
				//conn.Close()
				return
			} else {
				fmt.Println("Error receiving:", err.Error())
				return
			}
			
		} else {
			err = conn.SetReadDeadline(time.Now().Add(1*time.Second))
			if err != nil {
				fmt.Println("Could not set deadline")
			}

			fmt.Printf("Received %d bytes from %s: %s\n", n, remoteAddr, string(buffer[:n]))

			str := string(buffer[:n])

			str = strings.TrimSpace(str)

			words := strings.Fields(str)

			if len(words) > 0 {
				lastWord := words[len(words)-1]
				rec,strerr := strconv.Atoi(lastWord)
				if strerr != nil {
					fmt.Println("can not convert to string")
				}
				recieved <- rec
				
			} else {
				fmt.Println("No words found in the string")
			}

			
			
		}
		

		// Process data (for now, just print it)
		// Add your processing code here
	}
}


func main() {
	var count int
	next_start := 0

	//myState := 2
	errch := make(chan int)
	recieved := make(chan int)

	go udp_recieve(errch, recieved)

	Loop:
	for{
		select {
		case a := <- recieved:
			next_start = a
			fmt.Println(next_start)
		case a := <- errch:
			fmt.Println(a)
			break Loop
			//myState = 1
		}
	}
	
	fmt.Println("Starts program B")
	cmd := exec.Command("gnome-terminal", "--","go", "run", "../Exercise4/main.go")
	err := cmd.Start()
	if err != nil{
		fmt.Printf("Command finished with error: %v", err)
	}

	conn := dial_UDP()

	count = next_start
	for {
		fmt.Println(count)
		udp_send(count, conn)
		time.Sleep(500*time.Millisecond)
		count++
	}


	//for i := 0; i < 10; i++{
	//	fmt.Println(i)
	//	time.Sleep(1*time.Second)
	//}
}