// backend/internal/infrastructure/storage/minio.go

package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

const defaultOpTimeout = 30 * time.Second

var httpClient = &http.Client{Timeout: 30 * time.Second}

func opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultOpTimeout)
}

type cancelReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelReadCloser) Close() error {
	defer c.cancel()
	return c.ReadCloser.Close()
}

type MinIOClientInterface interface {
	UploadAvatar(userID string, file *multipart.FileHeader) (string, error)
	DeleteAvatar(fileName string) error
	GetAvatarURL(fileName string) (string, error)
	GetAvatarFile(fileName string) (io.ReadCloser, error)
	GetAvatarContentType(fileName string) (string, error)
	GetAvatarObjectInfo(fileName string) (minio.ObjectInfo, error)
	CheckBucketExists() error
	GetBucketInfo() (map[string]interface{}, error)

	UploadFlag(countryCode string, svgContent []byte) (string, error)
	UploadFlagFromURL(countryCode string, url string) (string, error)
	GetFlagFile(countryCode string) (io.ReadCloser, error)
	GetFlagURL(countryCode string) (string, error)
	DeleteFlag(countryCode string) error
	FlagExists(countryCode string) (bool, error)

	UploadFlagPNGFromURL(countryCode string, format string, url string) (string, error)
	GetFlagPNGFile(countryCode string, format string) (io.ReadCloser, error)
}

type MinIOClient struct {
	client     *minio.Client
	bucketName string
	logger     *zap.Logger
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucketName string, useSSL bool, logger *zap.Logger) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	service := &MinIOClient{
		client:     client,
		bucketName: bucketName,
		logger:     logger,
	}

	if err := service.createBucketIfNotExists(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *MinIOClient) Client() *minio.Client {
	return s.client
}

func (s *MinIOClient) BucketName() string {
	return s.bucketName
}

func (s *MinIOClient) createBucketIfNotExists() error {
	ctx, cancel := opCtx()
	defer cancel()

	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MinIOClient) UploadAvatar(userID string, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))

	fileName := fmt.Sprintf("avatar/%s", userID)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		if err := src.Close(); err != nil {
			s.logger.Error("Error closing file", zap.Error(err))
		}
	}()

	contentType := getContentType(ext)

	ctx, cancel := opCtx()
	defer cancel()

	_, err = s.client.PutObject(ctx, s.bucketName, fileName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func (s *MinIOClient) DeleteAvatar(fileName string) error {
	if fileName == "" {
		return nil
	}

	ctx, cancel := opCtx()
	defer cancel()

	return s.client.RemoveObject(ctx, s.bucketName, fileName, minio.RemoveObjectOptions{})
}

func (s *MinIOClient) GetAvatarURL(fileName string) (string, error) {
	if fileName == "" {
		return "", nil
	}

	ctx, cancel := opCtx()
	defer cancel()

	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, fileName, 24*time.Hour, nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func (s *MinIOClient) GetAvatarFile(fileName string) (io.ReadCloser, error) {
	ctx, cancel := opCtx()
	obj, err := s.client.GetObject(ctx, s.bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		cancel()
		return nil, err
	}
	return &cancelReadCloser{ReadCloser: obj, cancel: cancel}, nil
}

func (s *MinIOClient) GetAvatarContentType(fileName string) (string, error) {
	ctx, cancel := opCtx()
	defer cancel()

	objInfo, err := s.client.StatObject(ctx, s.bucketName, fileName, minio.StatObjectOptions{})
	if err != nil {
		return "", err
	}
	return objInfo.ContentType, nil
}

func (s *MinIOClient) GetAvatarObjectInfo(fileName string) (minio.ObjectInfo, error) {
	ctx, cancel := opCtx()
	defer cancel()

	objInfo, err := s.client.StatObject(ctx, s.bucketName, fileName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, err
	}
	return objInfo, nil
}

func (s *MinIOClient) CheckBucketExists() error {
	ctx, cancel := opCtx()
	defer cancel()

	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("bucket does not exist")
	}
	return nil
}

func (s *MinIOClient) GetBucketInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	details := make(map[string]interface{})

	details["bucket_name"] = s.bucketName

	location, err := s.client.GetBucketLocation(ctx, s.bucketName)
	if err != nil {
		return nil, err
	}
	details["location"] = location

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{})

	var totalObjects int64
	var totalSize int64
	var lastModified time.Time
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		totalObjects++
		totalSize += object.Size
		if object.LastModified.After(lastModified) {
			lastModified = object.LastModified
		}
	}

	details["total_objects"] = totalObjects
	details["total_size_bytes"] = totalSize
	details["total_size_human"] = formatBytes(totalSize)
	if totalObjects > 0 && !lastModified.IsZero() {
		details["last_object_modified"] = lastModified.Format(time.RFC3339)
	}

	return details, nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

func (s *MinIOClient) getFlagPath(countryCode string) string {
	return fmt.Sprintf("flags-svg/%s.svg", strings.ToLower(countryCode))
}

func (s *MinIOClient) UploadFlag(countryCode string, svgContent []byte) (string, error) {
	if len(svgContent) == 0 {
		return "", fmt.Errorf("empty SVG content for country: %s", countryCode)
	}

	if !strings.Contains(string(svgContent), "<svg") {
		return "", fmt.Errorf("invalid SVG content for country: %s", countryCode)
	}

	flagPath := s.getFlagPath(countryCode)

	ctx, cancel := opCtx()
	defer cancel()

	_, err := s.client.PutObject(ctx, s.bucketName, flagPath, bytes.NewReader(svgContent), int64(len(svgContent)), minio.PutObjectOptions{
		ContentType: "image/svg+xml",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload flag for %s: %w", countryCode, err)
	}

	s.logger.Debug("Flag uploaded", zap.String("country", countryCode), zap.String("path", flagPath))
	return flagPath, nil
}

func (s *MinIOClient) UploadFlagFromURL(countryCode string, url string) (string, error) {
	flagPath := s.getFlagPath(countryCode)

	existCtx, existCancel := opCtx()
	defer existCancel()

	_, err := s.client.StatObject(existCtx, s.bucketName, flagPath, minio.StatObjectOptions{})
	if err == nil {
		s.logger.Debug("Flag already exists, skipping", zap.String("country", countryCode))
		return flagPath, nil
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download flag for %s: %w", countryCode, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Error("Error closing response body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download flag for %s: status %d", countryCode, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read flag content for %s: %w", countryCode, err)
	}

	return s.UploadFlag(countryCode, content)
}

func (s *MinIOClient) GetFlagFile(countryCode string) (io.ReadCloser, error) {
	flagPath := s.getFlagPath(countryCode)
	ctx, cancel := opCtx()
	obj, err := s.client.GetObject(ctx, s.bucketName, flagPath, minio.GetObjectOptions{})
	if err != nil {
		cancel()
		return nil, err
	}
	return &cancelReadCloser{ReadCloser: obj, cancel: cancel}, nil
}

func (s *MinIOClient) GetFlagURL(countryCode string) (string, error) {
	flagPath := s.getFlagPath(countryCode)

	ctx, cancel := opCtx()
	defer cancel()

	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, flagPath, 24*time.Hour, nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func (s *MinIOClient) DeleteFlag(countryCode string) error {
	flagPath := s.getFlagPath(countryCode)

	ctx, cancel := opCtx()
	defer cancel()

	return s.client.RemoveObject(ctx, s.bucketName, flagPath, minio.RemoveObjectOptions{})
}

func (s *MinIOClient) FlagExists(countryCode string) (bool, error) {
	flagPath := s.getFlagPath(countryCode)

	ctx, cancel := opCtx()
	defer cancel()

	_, err := s.client.StatObject(ctx, s.bucketName, flagPath, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *MinIOClient) getFlagPNGPath(countryCode string, format string) string {
	return fmt.Sprintf("flags-png-%s/%s.png", format, strings.ToLower(countryCode))
}

func (s *MinIOClient) GetFlagPNGFile(countryCode string, format string) (io.ReadCloser, error) {
	flagPath := s.getFlagPNGPath(countryCode, format)
	ctx, cancel := opCtx()
	obj, err := s.client.GetObject(ctx, s.bucketName, flagPath, minio.GetObjectOptions{})
	if err != nil {
		cancel()
		return nil, err
	}
	return &cancelReadCloser{ReadCloser: obj, cancel: cancel}, nil
}

func (s *MinIOClient) UploadFlagPNGFromURL(countryCode string, format string, url string) (string, error) {
	flagPath := s.getFlagPNGPath(countryCode, format)

	existCtx, existCancel := opCtx()
	defer existCancel()

	_, err := s.client.StatObject(existCtx, s.bucketName, flagPath, minio.StatObjectOptions{})
	if err == nil {
		s.logger.Debug("PNG flag already exists, skipping",
			zap.String("country", countryCode),
			zap.String("format", format),
		)
		return flagPath, nil
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download PNG flag for %s (%spx): %w", countryCode, format, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Error("Error closing response body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download PNG flag for %s (%spx): status %d", countryCode, format, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read PNG flag content for %s (%spx): %w", countryCode, format, err)
	}

	if len(content) == 0 {
		return "", fmt.Errorf("empty PNG content for country: %s (%spx)", countryCode, format)
	}

	uploadCtx, uploadCancel := opCtx()
	defer uploadCancel()

	_, err = s.client.PutObject(uploadCtx, s.bucketName, flagPath, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType: "image/png",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload PNG flag for %s (%spx): %w", countryCode, format, err)
	}

	s.logger.Debug("PNG flag uploaded",
		zap.String("country", countryCode),
		zap.String("format", format),
		zap.String("path", flagPath),
	)
	return flagPath, nil
}
