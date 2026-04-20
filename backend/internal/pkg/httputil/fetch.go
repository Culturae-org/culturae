// backend/internal/pkg/httputil/fetch.go

package httputil

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FetchResponse struct {
	Body  string
	Error error
}

func FetchURL(url string) FetchResponse {
	resp, err := http.Get(url)
	if err != nil {
		return FetchResponse{Error: err}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return FetchResponse{Error: fmt.Errorf("failed to get URL %s: status code %d", url, resp.StatusCode)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FetchResponse{Error: err}
	}

	return FetchResponse{Body: string(body)}
}

type ChecksumMismatchError struct {
	URL      string
	Expected string
	Actual   string
}

func (e *ChecksumMismatchError) Error() string {
	return fmt.Sprintf("checksum mismatch for %s: expected %s, got %s", e.URL, e.Expected, e.Actual)
}

func IsChecksumMismatch(err error) (*ChecksumMismatchError, bool) {
	if err == nil {
		return nil, false
	}
	cm := &ChecksumMismatchError{}
	if errors.As(err, &cm) {
		return cm, true
	}
	return nil, false
}

func FetchURLWithChecksum(url string, expectedChecksum string) FetchResponse {
	resp, err := http.Get(url)
	if err != nil {
		return FetchResponse{Error: err}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return FetchResponse{Error: fmt.Errorf("failed to get URL %s: status code %d", url, resp.StatusCode)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FetchResponse{Error: err}
	}

	if expectedChecksum != "" {
		hash := sha256.Sum256(body)
		actualChecksum := hex.EncodeToString(hash[:])
		expected := strings.TrimPrefix(expectedChecksum, "sha256-")
		if actualChecksum != expected {
			return FetchResponse{
				Error: &ChecksumMismatchError{
					URL:      url,
					Expected: expected,
					Actual:   actualChecksum,
				},
			}
		}
	}

	return FetchResponse{Body: string(body)}
}
