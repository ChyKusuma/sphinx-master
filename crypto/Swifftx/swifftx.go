package swifftx

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lSHA3
#include <stdlib.h>
#include <string.h>
#include <stdio.h>  // Include for sprintf
#include "SHA3.h"

// Function to hash input message using SWIFFTX
void HashInput(const char *input, int length, char *output) {
    BitSequence resultingDigest[SWIFFTX_OUTPUT_BLOCK_SIZE] = {0};
    HashReturn exitCode;

    exitCode = Hash(512, input, length * 8, resultingDigest);  // 512-bit output

    if (exitCode == SUCCESS) {
        for (int i = 0; i < 64; i++) { // 64 bytes for 512 bits
            sprintf(output + (i * 2), "%02X", resultingDigest[i]); // Convert to hex
        }
    }
}
*/
import "C"
import (
	"unsafe"
)

func SWIFFTXHash(input string) (string, error) {
	length := len(input)
	output := make([]byte, 128) // 64 bytes = 512 bits, each byte represented by 2 hex chars

	cInput := C.CString(input)
	defer C.free(unsafe.Pointer(cInput))

	C.HashInput(cInput, C.int(length), (*C.char)(unsafe.Pointer(&output[0])))

	return string(output), nil
}
