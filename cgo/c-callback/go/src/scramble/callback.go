package main

// #include "libcallback.h"
// Result bridge_func(callback f, char* msg)
// {
//		return f(msg);
// }
import "C"

// Not in the main file because using //export in a file places a restriction on the preamble:
// it must not contain any definitions, only declarations.

// executeCallback executes the callback on the message
func executeCallback(f C.callback, msg string) (string, C.Result) {
	cMsg := C.CString(msg)
	result := C.bridge_func(f, cMsg)
	return C.GoString(cMsg), result
}
