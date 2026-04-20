// backend/internal/pkg/identifier/publicid.go

package identifier

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/Culturae-org/culturae/internal/model"
)

func GeneratePublicID() string {
	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	part1 := fmt.Sprintf("%04d", int(bytes[0])<<8|int(bytes[1]))
	part2 := fmt.Sprintf("%04d", int(bytes[2])<<8|int(bytes[3]))
	return fmt.Sprintf("%s-%s", part1, part2)
}

type PublicIDChecker interface {
	GetByPublicID(publicID string) (*model.User, error)
}

func GenerateUniquePublicID(checker PublicIDChecker) (string, error) {
	for {
		publicID := GeneratePublicID()
		_, err := checker.GetByPublicID(publicID)
		if errors.Is(err, model.ErrUserNotFound) {
			return publicID, nil
		}
		if err != nil {
			return "", err
		}
	}
}
