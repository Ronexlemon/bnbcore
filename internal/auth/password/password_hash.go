package password

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const separator = "|" // bcrypt hashes never contain |

type PasswordHasher struct {
	peppers        map[string]string
	activePepperID string
}

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

	// "v1|$2a$10$..."  — pipe separator is safe, bcrypt never uses it
	return fmt.Sprintf("%s%s%s", p.activePepperID, separator, string(hashedBytes)), nil
}

func (p *PasswordHasher) Compare(password, storedHash string) (bool, bool) {
	parts := strings.SplitN(storedHash, separator, 2)
	if len(parts) != 2 {
		return false, false // malformed hash
	}

	pepperID := parts[0]
	bcryptHash := parts[1] // now the full "$2a$10$..." is preserved

	pepper, exists := p.peppers[pepperID]
	if !exists {
		return false, false // pepper key deleted from config
	}

	err := bcrypt.CompareHashAndPassword([]byte(bcryptHash), []byte(password+pepper))
	if err != nil {
		return false, false
	}

	needsUpgrade := pepperID != p.activePepperID
	return true, needsUpgrade
}