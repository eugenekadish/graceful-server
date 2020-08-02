package main

// #cgo LDFLAGS: -L/home/eugene/Documents/miscelaneous/graceful-server/build/src -l:main -lstdc++ -lgmp -lgomp
// #include "qap.h"
//
// #include <stdio.h>
// #include <stdlib.h>
//
// static void myprint(char* s) {
//   printf("%s\n", s);
// }
import "C"
import "unsafe"

func main() {

	C.init()
	cs := C.CString("Hello from stdio")
	C.myprint(cs)
	C.free(unsafe.Pointer(cs))

}
