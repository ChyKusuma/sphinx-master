// MIT License
//
// Copyright (c) 2024 sphinx-core
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package hashtree

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"syscall"

	"github.com/syndtr/goleveldb/leveldb"
)

var maxFileSize = 1 << 30 // 1 GiB max file size for memory mapping

// HashTreeNode represents a node in the hash tree
type HashTreeNode struct {
	Hash  []byte        `json:"hash"`            // Hash of the node's data
	Left  *HashTreeNode `json:"left,omitempty"`  // Left child node
	Right *HashTreeNode `json:"right,omitempty"` // Right child node
}

// Compute the hash of a given data slice
func computeHash(data []byte) []byte {
	hash := sha256.Sum256(data) // Compute SHA-256 hash
	return hash[:]
}

// BuildHashTree builds a Merkle hash tree from leaf nodes.
// It returns the root node of the hash tree, which is computed by repeatedly
// combining and hashing pairs of leaf nodes and intermediate nodes.
func BuildHashTree(leaves [][]byte) *HashTreeNode {
	// Create an array of hash tree nodes, where each leaf node is hashed.
	nodes := make([]*HashTreeNode, len(leaves))
	for i, leaf := range leaves {
		// Create a new leaf node by computing the hash of each leaf and storing it in the node.
		nodes[i] = &HashTreeNode{Hash: computeHash(leaf)}
	}

	// Continue building the tree until there is only one node left, the root.
	for len(nodes) > 1 {
		// Prepare the next level of the tree.
		var nextLevel []*HashTreeNode

		// Iterate over the current level two nodes at a time.
		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				// Combine the hashes of two sibling nodes (left and right).
				left, right := nodes[i], nodes[i+1]
				// Concatenate the two hashes and compute the hash of the result to create the parent node.
				hash := computeHash(append(left.Hash, right.Hash...))
				// Append the new parent node to the next level, storing references to its children.
				nextLevel = append(nextLevel, &HashTreeNode{Hash: hash, Left: left, Right: right})
			} else {
				// If there is an odd number of nodes, carry the last node to the next level.
				nextLevel = append(nextLevel, nodes[i])
			}
		}

		// Move up one level, using the newly created nodes as the current level.
		nodes = nextLevel
	}

	// Return the single remaining node, which is the root of the hash tree.
	return nodes[0]
}

// Generate random data of specified length
func GenerateRandomData(size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data) // Fill with random data
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Save root hash to file
func SaveRootHashToFile(root *HashTreeNode, filename string) error {
	return ioutil.WriteFile(filename, root.Hash, 0644) // Save root hash to file
}

// Load root hash from file
func LoadRootHashFromFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename) // Read root hash from file
}

// SaveLeavesToDB saves leaf node data to LevelDB.
// The function takes a slice of leaf data (leaves) and stores each leaf in the database (db).
func SaveLeavesToDB(db *leveldb.DB, leaves [][]byte) error {
	// Iterate over the leaves to be saved to the database
	for i, leaf := range leaves {
		// Generate a unique key for each leaf using a formatted string with its index
		key := fmt.Sprintf("leaf-%d", i)
		// Store the leaf node in LevelDB using the generated key
		err := db.Put([]byte(key), leaf, nil) // Insert the leaf node into the database
		// If an error occurs while saving the leaf, return the error
		if err != nil {
			return err // Return the error to the caller
		}
	}
	// Return nil to indicate that all leaf nodes were saved successfully
	return nil
}

// Fetch leaf from LevelDB
func FetchLeafFromDB(db *leveldb.DB, key string) ([]byte, error) {
	return db.Get([]byte(key), nil) // Retrieve leaf node from LevelDB
}

// Print the root hash of the hash tree
func PrintRootHash(root *HashTreeNode) {
	fmt.Printf("Root Hash: %x\n", root.Hash) // Print root hash
}

// PruneOldLeaves removes old leaf nodes from the LevelDB.
// It takes a specified number of leaves (numLeaves) and deletes them by key from the database.
func PruneOldLeaves(db *leveldb.DB, numLeaves int) error {
	// Loop over the number of leaves to be deleted
	for i := 0; i < numLeaves; i++ {
		// Generate the key for the leaf node using a formatted string
		key := fmt.Sprintf("leaf-%d", i)
		// Attempt to delete the leaf node by key
		err := db.Delete([]byte(key), nil) // Remove old leaf node
		// If an error occurs, return it, except for the ErrNotFound case (ignore if key not found)
		if err != nil && err != leveldb.ErrNotFound {
			return err // Return any error other than 'not found'
		}
	}
	// Return nil if the operation completes successfully without errors
	return nil
}

// SaveLeavesBatchToDB performs batch operations for LevelDB to save leaf nodes efficiently.
// Using a batch operation improves performance by reducing the number of write calls to the database.
func SaveLeavesBatchToDB(db *leveldb.DB, leaves [][]byte) error {
	// Create a new batch to accumulate multiple write operations
	batch := new(leveldb.Batch)
	// Iterate over the leaves to be added
	for i, leaf := range leaves {
		// Generate the key for each leaf node using a formatted string
		key := fmt.Sprintf("leaf-%d", i)
		// Add the leaf node to the batch
		batch.Put([]byte(key), leaf) // Queue the leaf for batch write
	}
	// Execute the batch write to LevelDB, applying all queued operations at once
	return db.Write(batch, nil) // Write the batch to the database
}

// FetchLeafConcurrent retrieves a leaf node from LevelDB while ensuring it handles concurrent access safely.
// In this example, concurrency is handled implicitly by the LevelDB API, which can manage simultaneous read operations.
func FetchLeafConcurrent(db *leveldb.DB, key string) ([]byte, error) {
	// Retrieve the leaf node from LevelDB using its key
	return db.Get([]byte(key), nil) // Fetch the leaf data
}

// setMaxFileSize updates the global maxFileSize variable based on the provided size in GiB (gibibytes).
// This function ensures the size is valid and converts it to bytes for use in file size limits.
func setMaxFileSize(sizeInGiB int) {
	// Check if the size is greater than 0 to ensure a valid file size is provided
	if sizeInGiB <= 0 {
		// Print an error message and return if the size is invalid
		fmt.Println("Invalid size. Must be greater than 0.")
		return
	}
	// Convert the provided size from GiB to bytes (1 GiB = 2^30 bytes) and set the global maxFileSize
	maxFileSize = sizeInGiB * (1 << 30) // Convert GiB to bytes
}

// MemoryMapFile maps a file into memory with size checks
func MemoryMapFile(filename string) ([]byte, error) {
	// Open the file for reading using os.Open
	file, err := os.Open(filename)
	if err != nil {
		// Return an error if the file cannot be opened, providing context
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	// Ensure the file is closed when the function returns, even if an error occurs
	defer file.Close()

	// Get file statistics, such as size, using file.Stat
	stat, err := file.Stat()
	if err != nil {
		// Return an error if there is an issue retrieving file stats
		return nil, fmt.Errorf("error getting file stats: %w", err)
	}

	// Get the size of the file from the stat object
	size := stat.Size()
	// Compare the file size with the maximum allowed file size (maxFileSize)
	// Convert maxFileSize to int64 for proper comparison, since file size is in int64
	if size > int64(maxFileSize) {
		// Return an error if the file size exceeds the allowed limit
		return nil, fmt.Errorf("file size exceeds maximum limit of %d bytes", maxFileSize)
	}

	// Memory-map the file using syscall.Mmap
	// Arguments: file descriptor (int(file.Fd())), offset (0), size, protection (read-only), shared mapping
	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		// Return an error if memory mapping fails
		return nil, fmt.Errorf("error mapping file: %w", err)
	}

	// Return the memory-mapped data as a byte slice
	return data, nil
}

// UnmapFile unmaps a file from memory with error handling
func UnmapFile(data []byte) error {
	if err := syscall.Munmap(data); err != nil {
		return fmt.Errorf("error unmapping file: %w", err)
	}
	return nil
}

// Concurrency control for memory-mapped file access
var mu sync.Mutex

// SafeMemoryMapFile safely memory-maps a file by ensuring exclusive access using a mutex lock
func SafeMemoryMapFile(filename string) ([]byte, error) {
	// Lock the mutex to prevent concurrent access to the file mapping
	mu.Lock()
	// Ensure the mutex is unlocked when the function completes (either successfully or with an error)
	defer mu.Unlock()
	// Call the actual MemoryMapFile function to memory-map the specified file
	return MemoryMapFile(filename)
}

// SafeUnmapFile safely unmaps a file by ensuring exclusive access using a mutex lock
func SafeUnmapFile(data []byte) error {
	// Lock the mutex to prevent concurrent access to the unmapping process
	mu.Lock()
	// Ensure the mutex is unlocked when the function completes (either successfully or with an error)
	defer mu.Unlock()
	// Call the actual UnmapFile function to unmap the provided file data
	return UnmapFile(data)
}
