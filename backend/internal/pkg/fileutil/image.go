// backend/internal/pkg/fileutil/image.go

package fileutil

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
)

const MaxFileSize = 5 * 1024 * 1024

var (
	allowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}
	allowedImageExtensions = map[string]bool{
		".png":  true,
		".jpeg": true,
		".jpg":  true,
	}
)

func IsValidFileType(file *multipart.FileHeader) bool {
	contentType := file.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return false
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	return allowedImageExtensions[ext]
}

func IsValidFileTypeWithConfig(file *multipart.FileHeader, allowedMimes, allowedExts []string) bool {
	contentType := file.Header.Get("Content-Type")
	mimeOK := false
	for _, m := range allowedMimes {
		if m == contentType {
			mimeOK = true
			break
		}
	}
	if !mimeOK {
		return false
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	for _, e := range allowedExts {
		if e == ext {
			return true
		}
	}
	return false
}

func FormatAvatarURL(userID string) string {
	return fmt.Sprintf("avatar/%s", userID)
}

func FormatAvatarURLFromFullPath(fullPath string) string {
	if fullPath == "" {
		return ""
	}

	parts := strings.Split(fullPath, "/")
	if len(parts) >= 2 {
		return fmt.Sprintf("avatar/%s", parts[1])
	}

	return ""
}
