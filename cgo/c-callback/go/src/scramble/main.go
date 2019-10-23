package main

// #include "libcallback.h"
import "C"
import (
	"fmt"
)

// receive "receives" the initial message
func receive() string {
	msg := "the cat in the hat"
	fmt.Println("Received msg:", msg)
	return msg
}
// transmit "transmits" the modified message
func transmit(msg string)  {
	fmt.Println("Transmitted msg:", msg)
}

//export scramble_message
// scramble_message receives a message and uses the supplied C callback to modify that message.
// Finally the modified message is transmitted.
func scramble_message(f C.callback) C.Result {
	msg := receive()
	msg, result := executeCallback(f, msg)
	if result == C.Failure {
		fmt.Println("Callback failed")
		return result
	}
	transmit(msg)
	return C.Success
}

func main() { }
