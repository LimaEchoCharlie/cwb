# C callbacks in CGO

Code shows an example of how to produce a C API function that accepts a 
C callback function from Go code. See the [cgo](https://golang.org/cmd/cgo/) 
wiki for the original example.

To build:

    ./build.sh

To run: 

    env DYLD_LIBRARY_PATH=$(pwd)/build build/scrambler