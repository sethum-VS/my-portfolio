package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

const resumeObjectKey = "resumes/cv.pdf"

var storageClient *storage.Client

// InitStorage creates the GCS client (call once from main).
func InitStorage(ctx context.Context) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	storageClient = client
	return nil
}

func resumeBucket() (string, error) {
	bucket := os.Getenv("GCP_STORAGE_BUCKET")
	if bucket == "" {
		return "", fmt.Errorf("GCP_STORAGE_BUCKET is not set")
	}
	return bucket, nil
}

// UploadResumePDF uploads a PDF to GCS and returns the gs:// URI.
func UploadResumePDF(ctx context.Context, r io.Reader, contentType string) (string, error) {
	if storageClient == nil {
		return "", fmt.Errorf("storage client not initialized")
	}

	bucket, err := resumeBucket()
	if err != nil {
		return "", err
	}

	if contentType == "" {
		contentType = "application/pdf"
	}

	w := storageClient.Bucket(bucket).Object(resumeObjectKey).NewWriter(ctx)
	w.ContentType = contentType
	w.CacheControl = "private, max-age=3600"

	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return "", fmt.Errorf("failed to upload resume: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize upload: %w", err)
	}

	return fmt.Sprintf("gs://%s/%s", bucket, resumeObjectKey), nil
}

// DownloadResumePDF reads the PDF bytes from a gs:// URI.
func DownloadResumePDF(ctx context.Context, gsURI string) ([]byte, error) {
	if storageClient == nil {
		return nil, fmt.Errorf("storage client not initialized")
	}

	bucket, object, err := parseGSURI(gsURI)
	if err != nil {
		return nil, err
	}

	rc, err := storageClient.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open GCS object: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read GCS object: %w", err)
	}
	return data, nil
}

func parseGSURI(gsURI string) (bucket, object string, err error) {
	const prefix = "gs://"
	if !strings.HasPrefix(gsURI, prefix) {
		return "", "", fmt.Errorf("invalid gs URI: %s", gsURI)
	}
	path := strings.TrimPrefix(gsURI, prefix)
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid gs URI: %s", gsURI)
	}
	return parts[0], parts[1], nil
}
