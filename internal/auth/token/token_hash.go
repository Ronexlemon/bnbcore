package token



import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type TokenHasher struct {
	dataHashingKey []byte
}

// NewTokenHasher unlocks the underlying data hashing key using your rotatable environment master key
func NewTokenHasher(masterKeyHex string, encryptedDataKeyHex string) (*TokenHasher, error) {
	masterKey, err := hex.DecodeString(masterKeyHex)
	if err != nil {
		return nil, errors.New("failed to decode master configuration key hex")
	}

	encryptedDK, err := hex.DecodeString(encryptedDataKeyHex)
	if err != nil {
		return nil, errors.New("failed to decode encrypted data key payload hex")
	}

	if len(encryptedDK) < 12 {
		return nil, errors.New("encrypted data key payload is too short or malformed")
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Split the 12-byte nonce apart from the encrypted ciphertext payload
	nonce := encryptedDK[:12]
	ciphertext := encryptedDK[12:]

	staticDataKey, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("master key rotation mismatch: cannot decrypt the data key envelope")
	}

	return &TokenHasher{dataHashingKey: staticDataKey}, nil
}

// Hash guarantees identical, matching outputs for direct SQL indexing strings (equality matching)
func (t *TokenHasher) Hash(token string) string {
	mac := hmac.New(sha256.New, t.dataHashingKey)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}