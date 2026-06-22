package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type ObjectInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	LastModified time.Time `json:"last_modified"`
}

type Service struct {
	client       *minio.Client
	bucketPrefix string
}

func NewService(client *minio.Client, bucket string) *Service {
	return &Service{
		client:       client,
		bucketPrefix: bucket,
	}
}

func (s *Service) BucketNameForUser(userID string) string {
	name := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s-%s", s.bucketPrefix, userID)))
	builder := strings.Builder{}
	lastDash := false
	for _, r := range name {
		allowed := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if allowed {
			builder.WriteRune(r)
			lastDash = false
			continue
		}

		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}

	bucketName := strings.Trim(builder.String(), "-.")
	if len(bucketName) < 3 {
		bucketName = bucketName + "-img"
	}
	if len(bucketName) > 63 {
		bucketName = bucketName[:63]
		bucketName = strings.TrimRight(bucketName, "-.")
	}
	if len(bucketName) < 3 {
		bucketName = "img-bucket"
	}

	return bucketName
}

func (s *Service) EnsureUserBucket(ctx context.Context, userID string) (string, error) {
	bucketName := s.BucketNameForUser(userID)

	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return "", err
	}
	if exists {
		return bucketName, nil
	}

	if err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		return "", err
	}

	return bucketName, nil
}

func (s *Service) UploadUserObject(ctx context.Context, userID string, objectName string, reader io.Reader, size int64, contentType string) error {
	bucketName, err := s.EnsureUserBucket(ctx, userID)
	if err != nil {
		return err
	}

	_, err = s.client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Service) ListUserObjects(ctx context.Context, userID string) ([]ObjectInfo, error) {
	bucketName := s.BucketNameForUser(userID)
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []ObjectInfo{}, nil
	}

	objects := make([]ObjectInfo, 0)
	for object := range s.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true}) {
		if object.Err != nil {
			return nil, object.Err
		}

		contentType := object.ContentType
		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(object.Key))
		}

		objects = append(objects, ObjectInfo{
			Name:         object.Key,
			Size:         object.Size,
			ContentType:  contentType,
			LastModified: object.LastModified,
		})
	}

	return objects, nil
}

func (s *Service) StatUserObject(ctx context.Context, userID string, objectName string) (*ObjectInfo, error) {
	bucketName := s.BucketNameForUser(userID)
	info, err := s.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	contentType := info.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(objectName))
	}

	return &ObjectInfo{
		Name:         objectName,
		Size:         info.Size,
		ContentType:  contentType,
		LastModified: info.LastModified,
	}, nil
}

func (s *Service) ReadUserObject(ctx context.Context, userID string, objectName string) ([]byte, string, error) {
	bucketName := s.BucketNameForUser(userID)
	object, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	defer object.Close()

	info, err := object.Stat()
	if err != nil {
		return nil, "", err
	}

	contentType := info.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(objectName))
	}

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, "", err
	}

	return data, contentType, nil
}

func (s *Service) RemoveUserObject(ctx context.Context, userID string, objectName string) error {
	bucketName := s.BucketNameForUser(userID)
	return s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}
