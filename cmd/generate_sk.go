package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

func main() {
	sk, err := generateSymmetricKey(32)
	if err != nil {
		log.Fatal(err)
	}

	filePath := "sk.txt"
	if err = os.WriteFile(filePath, []byte(sk), 0644); err != nil {
		log.Fatalf("failed to save sk.txt: %v", err)
	}
	log.Println("saved sk.txt successfully")
}

func generateSymmetricKey(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
