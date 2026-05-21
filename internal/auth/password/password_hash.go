package password



import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct {
	peppers        map[string]string // e.g., {"v1": "old_pepper", "v2": "new_pepper"}
	activePepperID string            // e.g., "v2"
}

// NewPasswordHasher initializes the password manager with your system peppers
func NewPasswordHasher(peppers map[string]string, activePepperID string) (*PasswordHasher, error) {
	if _, exists := peppers[activePepperID]; !exists {
		return nil, errors.New("active pepper ID must exist in the provided peppers map")
	}
	return &PasswordHasher{
		peppers:        peppers,
		activePepperID: activePepperID,
	}, nil
}


func (p *PasswordHasher) Hash(password string) (string, error) {
	activePepper := p.peppers[p.activePepperID]
	
	payload := password + activePepper

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(payload), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Store with a version prefix so we know how to verify it later (e.g., "v2$2a$10...")
	return fmt.Sprintf("%s$%s", p.activePepperID, string(hashedBytes)), nil
}

// Compare checks the password and returns: (isValid, needsMigrationUpgrade)
func (p *PasswordHasher) Compare(password, storedHash string) (bool, bool) {
	parts := strings.SplitN(storedHash, "$", 2)
	if len(parts) != 2 {
		return false, false // Malformed hash structure
	}

	pepperID := parts[0]
	actualBcryptHash := parts[1]

	pepper, exists := p.peppers[pepperID]
	if !exists {
		// This happens if an old pepper key was completely deleted from your config environment
		return false, false 
	}

	err := bcrypt.CompareHashAndPassword([]byte(actualBcryptHash), []byte(password+pepper))
	if err != nil {
		return false, false // Wrong password
	}

	// Password is correct! Now check if it was hashed with an outdated pepper key version
	needsUpgrade := pepperID != p.activePepperID
	return true, needsUpgrade
}