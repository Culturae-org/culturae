// backend/internal/pkg/crypto/password.go

package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"golang.org/x/crypto/argon2"
)

const (
	passwordMinLength = 8
	passwordMaxLength = 128
)

func IsValidPassword(password string) bool {
	l := len(password)
	if l < passwordMinLength || l > passwordMaxLength {
		return false
	}
	var hasUpper, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case (c >= 33 && c <= 47) || (c >= 58 && c <= 64) || (c >= 91 && c <= 96) || (c >= 123 && c <= 126):
			hasSpecial = true
		}
	}
	return hasUpper && hasDigit && hasSpecial
}

func CheckPassword(password, encoded string) (bool, error) {
	if !strings.HasPrefix(encoded, "$argon2id$") {
		return false, errors.New("unsupported hash format")
	}
	return checkPasswordArgon2(password, encoded)
}

func HashPassword(password string, p *model.ArgonParams) (string, error) {
	if p == nil {
		p = model.GetArgonParams()
	}

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		p.Memory, p.Iterations, p.Parallelism, b64Salt, b64Hash), nil
}

func checkPasswordArgon2(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	keyLen := len(hash)
	if keyLen <= 0 || keyLen > 256 {
		return false, fmt.Errorf("invalid hash length: %d", keyLen)
	}

	newHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(keyLen))
	if subtle.ConstantTimeCompare(newHash, hash) != 1 {
		return false, nil
	}

	return true, nil
}

func GenerateTempPassword() string {
	return fmt.Sprintf("Temp%d!", time.Now().UnixNano()%100000)
}
