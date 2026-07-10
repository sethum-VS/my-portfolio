package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const resumeObjectKey = "cv.pdf" // Bucket is "resumes", so path is just "cv.pdf"

// UploadResumePDF uploads a PDF to Supabase Storage and returns the supabase:// URI.
func UploadResumePDF(ctx context.Context, r io.Reader, contentType string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || serviceKey == "" {
		return "", fmt.Errorf("SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY is not set")
	}

	if contentType == "" {
		contentType = "application/pdf"
	}

	// Upload via Supabase Storage API
	url := fmt.Sprintf("%s/storage/v1/object/resumes/%s", supabaseURL, resumeObjectKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, r)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload to Supabase Storage: status=%d body=%s", resp.StatusCode, string(body))
	}

	return fmt.Sprintf("supabase://resumes/%s", resumeObjectKey), nil
}

// DownloadResumePDF reads the PDF bytes from a supabase:// URI.
func DownloadResumePDF(ctx context.Context, storageURI string) ([]byte, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || serviceKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY is not set")
	}

	// Parse custom uri scheme e.g. "supabase://resumes/cv.pdf"
	path := strings.TrimPrefix(storageURI, "supabase://resumes/")

	// Using the authenticated object endpoint to download from private bucket
	url := fmt.Sprintf("%s/storage/v1/object/authenticated/resumes/%s", supabaseURL, path)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+serviceKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download from Supabase Storage: status=%d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
