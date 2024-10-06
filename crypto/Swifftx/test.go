package main

import (
	"fmt"
	"log"
	// Adjust the import path based on your Go module path
)

func main() {
	// Example input string
	input := "Hello, SWIFFTX!"

	// Call the SWIFFTXHash function
	hash, err := swifftx.SWIFFTXHash(input)
	if err != nil {
		log.Fatalf("Error hashing input: %v", err)
	}

	// Output the result
	fmt.Printf("Input: %s\nHash: %s\n", input, hash)
}
