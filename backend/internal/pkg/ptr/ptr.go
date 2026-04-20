// backend/internal/pkg/ptr/ptr.go

package ptr

func String(s string) *string {
	return &s
}

func DefaultString(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}
