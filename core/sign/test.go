package main

import (
	"fmt"
	"log"

	"github.com/kasperdi/SPHINCSPLUS-golang/parameters"
	"github.com/sphinx-core/sphinx-master/core/hashtree"
	"github.com/sphinx-core/sphinx-master/core/sign"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	// Initialize parameters for SHAKE256-robust with N = 32
	params := parameters.MakeSphincsPlusSHAKE256256fRobust(false)

	// Open LevelDB
	db, err := leveldb.OpenFile("leaves_db", nil)
	if err != nil {
		log.Fatal("Failed to open LevelDB:", err)
	}
	defer db.Close()

	// Initialize the SphincsManager with the LevelDB instance
	manager := sign.NewSphincsManager(db)

	// Generate keys
	sk, pk := manager.GenerateKeys(params)

	// Serialize the secret key to bytes
	skBytes, err := manager.SerializeSK(sk)
	if err != nil {
		log.Fatal("Failed to serialize SK:", err)
	}
	fmt.Printf("Secret Key (SK): %x\n", skBytes)
	fmt.Printf("Size of Serialized SK: %d bytes\n", len(skBytes))

	// Serialize the public key to bytes
	pkBytes, err := manager.SerializePK(pk)
	if err != nil {
		log.Fatal("Failed to serialize PK:", err)
	}
	fmt.Printf("Public Key (PK): %x\n", pkBytes)
	fmt.Printf("Size of Serialized PK: %d bytes\n", len(pkBytes))

	// Sign a message
	message := []byte("Hello, world!")
	sig, merkleRoot, err := manager.SignMessage(params, message, sk)
	if err != nil {
		log.Fatal("Failed to sign message:", err)
	}

	// Serialize the signature to bytes
	sigBytes, err := manager.SerializeSignature(sig)
	if err != nil {
		log.Fatal("Failed to serialize signature:", err)
	}
	fmt.Printf("Signature: %x\n", sigBytes)
	fmt.Printf("Size of Serialized Signature: %d bytes\n", len(sigBytes))

	// Print Merkle Tree root hash and size
	fmt.Printf("Merkle Tree Root Hash: %x\n", merkleRoot.Hash)
	fmt.Printf("Size of Merkle Tree Root Hash: %d bytes\n", len(merkleRoot.Hash))

	// Save Merkle root hash to a file
	err = hashtree.SaveRootHashToFile(merkleRoot, "merkle_root_hash.bin")
	if err != nil {
		log.Fatal("Failed to save root hash to file:", err)
	}

	// Load Merkle root hash from a file
	loadedHash, err := hashtree.LoadRootHashFromFile("merkle_root_hash.bin")
	if err != nil {
		log.Fatal("Failed to load root hash from file:", err)
	}
	fmt.Printf("Loaded Merkle Tree Root Hash: %x\n", loadedHash)

	// Save leaves to LevelDB
	leaves := [][]byte{sigBytes} // Example usage
	err = hashtree.SaveLeavesToDB(db, leaves)
	if err != nil {
		log.Fatal("Failed to save leaves to DB:", err)
	}

	// Fetch a leaf from LevelDB
	leaf, err := hashtree.FetchLeafFromDB(db, "leaf-0")
	if err != nil {
		log.Fatal("Failed to fetch leaf from DB:", err)
	}
	fmt.Printf("Fetched Leaf: %x\n", leaf)

	// Call generateRandomData to make it used
	randomData, err := hashtree.GenerateRandomData(16)
	if err != nil {
		log.Fatal("Failed to generate random data:", err)
	}
	fmt.Printf("Random Data: %x\n", randomData)

	// Call printRootHash to make it used
	hashtree.PrintRootHash(merkleRoot)

	// Verify the signature and print the original message
	isValid := manager.VerifySignature(params, message, sig, pk, merkleRoot)
	fmt.Printf("Signature valid: %v\n", isValid)
	if isValid {
		fmt.Printf("Original Message: %s\n", message)
	}
}
