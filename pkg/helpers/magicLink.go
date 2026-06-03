package helpers

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
)


func GenerateMagicLink() (rawToken string, tokenHash string, err error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", "", fmt.Errorf("failed to generate secure random bytes: %w", err)
    }
    
    rawToken = hex.EncodeToString(b)
    tokenHash = HashToken(rawToken)
    
    return rawToken, tokenHash, nil
}

func HashToken(raw string) string {
    h := sha256.Sum256([]byte(raw))
    return hex.EncodeToString(h[:])
}