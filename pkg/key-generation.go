package main


import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func main() {
	//  Generate a brand new, random 32-byte Master Key
	masterKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, masterKey); err != nil {
		fmt.Printf("Error generating master key: %v\n", err)
		os.Exit(1)
	}

	// Generate the permanent 32-byte internal Data Hashing Key 
	staticDataKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, staticDataKey); err != nil {
		fmt.Printf("Error generating static data key: %v\n", err)
		os.Exit(1)
	}

	//  Encrypt the internal key using AES-GCM with the Master Key
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		fmt.Printf("Error creating cipher block: %v\n", err)
		os.Exit(1)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Printf("Error creating GCM: %v\n", err)
		os.Exit(1)
	}

	// Generate a secure random 12-byte nonce for AES-GCM
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Printf("Error generating nonce: %v\n", err)
		os.Exit(1)
	}

	// Seal encrypts the data key and appends the result to our nonce prefix
	encryptedPayload := aesGCM.Seal(nonce, nonce, staticDataKey, nil)

	// 4. Output formatting for your .env file
	fmt.Println("\n========================================================")
	fmt.Println("   COPY AND PASTE THESE VALUES DIRECTLY INTO YOUR .ENV  ")
	fmt.Println("========================================================\n")
	fmt.Printf("MasterKeyHex=\"%s\"\n", hex.EncodeToString(masterKey))
	fmt.Printf("EncryptedDataKeyHex=\"%s\"\n\n", hex.EncodeToString(encryptedPayload))
	fmt.Println("========================================================")
}