package main

import(
	"fmt"
	"net"
	"time"
)


func tcp_client() *net.TCPConn {
	// Define the server address and port
	
	/*addr := net.TCPAddr{
		Port: 20014,
		IP:   net.ParseIP("0.0.0.0"),
	}*/
	
	addr, err := net.ResolveTCPAddr("tcp", "10.100.23.129:33546")
	// Create a TCP socket
	conn, err := net.DialTCP("tcp",nil, addr)
	if err != nil {
		fmt.Println("Error dialing:", err.Error())
		return nil
	}
	//defer conn.Close()
	fmt.Printf("Connecting to %s\n", addr.String())
	return conn

}

func tcp_server() *net.TCPConn{
	addr, err := net.ResolveTCPAddr("tcp", "10.100.23.24:20014")
	// Create a TCP socket
	listener, err := net.ListenTCP("tcp",addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return nil
	}
	for {
	 conn, err := listener.AcceptTCP()

	 if err != nil {
		fmt.Println("Error accepting:", err.Error())
		return nil
		}

	 fmt.Printf("Connecting to %s\n", addr.String())
	 return conn
	}
	//defer conn.Close()
	
}

func tcp_recieve(conn *net.TCPConn){
	// Buffer to store received data
	buffer := make([]byte, 1024)

	for {
		// Receive data
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error receiving:", err.Error())
			continue
		}
		fmt.Printf("Received %d bytes: %s\n", n, string(buffer[:n]))

		

		// Process data (for now, just print it)
		// Add your processing code here
	}
}

func tcp_send(s string, conn *net.TCPConn){

	// Buffer to store received data
	//buffer := make([]byte, 0)
	//buffer = append(buffer, 7)

		// Receive data

		
	n, err := conn.Write([]byte(s))
	if err != nil {
		fmt.Println("Error sending:", err.Error())
		return
	}
	fmt.Printf("sent %d bytes\n",n)

	time.Sleep(2*time.Second)

		// Process data (for now, just print it)
		// Add your processing code here
	
}

func main(){

	connect := tcp_client()
	//tcp_client()
	time.Sleep(1*time.Second)
	defer connect.Close()
	go tcp_recieve(connect)
	
	go tcp_send("Connect to: 10.100.23.24:20014\000",connect)

	conn := tcp_server()
	time.Sleep(1*time.Second)
	defer conn.Close()

	go tcp_recieve(conn)
	go tcp_send("Julie sier hei\000",conn)

	for {

	}
	//time.Sleep(5*time.Second)
	
}