package images

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"myproj/internal/ai"
	"myproj/internal/storage"
)

type Service struct {
	repository          *Repository
	storage             *storage.Service
	ai                  *ai.Service
	maxUploadSizeByte   int64
	allowedContentTypes map[string]struct{}
}

type UploadInput struct {
	UserID      string
	Filename    string
	ContentType string
	Size        int64
	Reader      io.Reader
}

type ImportInput struct {
	UserID     string
	ObjectName string
}

func NewService(
	repository *Repository,
	storageService *storage.Service,
	aiService *ai.Service,
	maxUploadSizeBytes int64,
	allowedContentTypes []string,
) *Service {
	allowed := make(map[string]struct{}, len(allowedContentTypes))
	for _, contentType := range allowedContentTypes {
		normalized := strings.TrimSpace(strings.ToLower(contentType))
		if normalized == "" {
			continue
		}
		allowed[normalized] = struct{}{}
	}

	return &Service{
		repository:          repository,
		storage:             storageService,
		ai:                  aiService,
		maxUploadSizeByte:   maxUploadSizeBytes,
		allowedContentTypes: allowed,
	}
}

func (s *Service) UploadAndDescribe(ctx context.Context, input UploadInput) (*Image, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if input.Reader == nil {
		return nil, fmt.Errorf("file reader is required")
	}

	limitedReader := io.LimitReader(input.Reader, s.maxUploadSizeByte+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > s.maxUploadSizeByte {
		return nil, fmt.Errorf("file exceeds max upload size of %d bytes", s.maxUploadSizeByte)
	}

	contentType, err := s.detectAndValidateContentType(input.ContentType, data)
	if err != nil {
		return nil, err
	}

	objectName := generateObjectName(input.Filename)
	if err := s.storage.UploadUserObject(ctx, input.UserID, objectName, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
		return nil, err
	}

	description, err := s.ai.DescribeImage(ctx, contentType, data)
	if err != nil {
		return nil, err
	}

	image := &Image{
		UserID:           input.UserID,
		ObjectName:       objectName,
		OriginalFilename: fallbackFilename(input.Filename, objectName),
		ContentType:      contentType,
		SizeBytes:        int64(len(data)),
		Description:      description,
	}

	if err := s.repository.Create(ctx, image); err != nil {
		return nil, err
	}

	return image, nil
}

func (s *Service) ImportFromBucket(ctx context.Context, input ImportInput) (*Image, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if strings.TrimSpace(input.ObjectName) == "" {
		return nil, fmt.Errorf("object name is required")
	}

	if existing, err := s.repository.GetByObjectNameForUser(ctx, input.UserID, input.ObjectName); err == nil {
		if strings.TrimSpace(existing.Description) != "" {
			return existing, nil
		}
	} else if err != pgx.ErrNoRows {
		return nil, err
	}

	objectInfo, err := s.storage.StatUserObject(ctx, input.UserID, input.ObjectName)
	if err != nil {
		return nil, err
	}

	data, contentType, err := s.storage.ReadUserObject(ctx, input.UserID, input.ObjectName)
	if err != nil {
		return nil, err
	}

	contentType, err = s.detectAndValidateContentType(objectInfo.ContentType, data)
	if err != nil {
		return nil, err
	}

	description, err := s.ai.DescribeImage(ctx, contentType, data)
	if err != nil {
		return nil, err
	}

	if existing, err := s.repository.GetByObjectNameForUser(ctx, input.UserID, input.ObjectName); err == nil {
		existing.Description = description
		if err := s.repository.UpdateDescription(ctx, existing.ID, input.UserID, description); err != nil {
			return nil, err
		}
		return existing, nil
	} else if err != pgx.ErrNoRows {
		return nil, err
	}

	image := &Image{
		UserID:           input.UserID,
		ObjectName:       input.ObjectName,
		OriginalFilename: path.Base(input.ObjectName),
		ContentType:      contentType,
		SizeBytes:        objectInfo.Size,
		Description:      description,
	}

	if err := s.repository.Create(ctx, image); err != nil {
		return nil, err
	}

	return image, nil
}

func (s *Service) ListImages(ctx context.Context, userID string) ([]Image, error) {
	return s.repository.ListByUserID(ctx, userID)
}

func (s *Service) GetImage(ctx context.Context, userID string, imageID string) (*Image, error) {
	return s.repository.GetByIDForUser(ctx, imageID, userID)
}

func (s *Service) DeleteImage(ctx context.Context, userID string, imageID string) error {
	image, err := s.repository.GetByIDForUser(ctx, imageID, userID)
	if err != nil {
		return err
	}

	if err := s.storage.RemoveUserObject(ctx, userID, image.ObjectName); err != nil {
		return err
	}

	return s.repository.DeleteByIDForUser(ctx, imageID, userID)
}

func (s *Service) ListBucketObjects(ctx context.Context, userID string) ([]storage.ObjectInfo, error) {
	objects, err := s.storage.ListUserObjects(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(s.allowedContentTypes) == 0 {
		return objects, nil
	}

	filtered := make([]storage.ObjectInfo, 0, len(objects))
	for _, object := range objects {
		if _, ok := s.allowedContentTypes[strings.ToLower(strings.TrimSpace(object.ContentType))]; ok {
			filtered = append(filtered, object)
		}
	}

	return filtered, nil
}

func (s *Service) detectAndValidateContentType(inputContentType string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("image data is empty")
	}

	detected := strings.ToLower(strings.TrimSpace(http.DetectContentType(data)))
	if detected == "application/octet-stream" {
		detected = strings.ToLower(strings.TrimSpace(inputContentType))
	}
	if detected == "" {
		return "", fmt.Errorf("could not determine image content type")
	}
	if !strings.HasPrefix(detected, "image/") {
		return "", fmt.Errorf("uploaded file must be an image")
	}
	if len(s.allowedContentTypes) > 0 {
		if _, ok := s.allowedContentTypes[detected]; !ok {
			return "", fmt.Errorf("unsupported image type %q", detected)
		}
	}

	return detected, nil
}

func generateObjectName(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return fmt.Sprintf("uploads/%s%s", uuid.NewString(), ext)
}

func fallbackFilename(filename string, objectName string) string {
	name := strings.TrimSpace(filename)
	if name != "" {
		return name
	}
	return path.Base(objectName)
}
