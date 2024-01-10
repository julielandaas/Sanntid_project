// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
  //  "time"
)

var i = 0

func incrementing(increment chan int, finish chan string) {
    
    //TODO: increment i 1000000 times
    for j := 0; j < 1000000; j++ {
        increment <- i
    }

    var finish_msg string = "worker done"
    finish <- finish_msg
}

func decrementing(decrement chan int, finish chan string) {
    //TODO: decrement i 1000000 times
    for j := 0; j < 1000001; j++ {
        decrement <- i
    }

    var finish_msg string = "worker done"
    finish <- finish_msg

}

func server(increment chan int, decrement chan int, read chan int) {
    
    for {
        select {
        case msg1:= <- increment:
            i = msg1 + 1

        case msg2:= <- decrement:
            i = msg2 - 1

        case read <- i:

        }
    }
}

func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    // It controls the maximum number of OS threads that are executing code simultaneously.
    runtime.GOMAXPROCS(2)  

	increment := make(chan int)
    decrement := make(chan int)
    read := make(chan int)
    finish := make(chan string)

    // TODO: Spawn both functions as goroutines
	go incrementing(increment, finish)

    go decrementing(decrement, finish)

    go server(increment, decrement, read)


    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    //time.Sleep(500*time.Millisecond)
    
    <- finish
    <- finish

    //var final_value int
    final_value := <- read



    Println("The magic number is:", final_value)
}
