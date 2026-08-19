package services

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sethum-VS/my-portfolio/internal/config"
)

func TestUploadResumePDF(t *testing.T) {
	var capturedUpsertHeader string
	var capturedAuthHeader string
	var capturedContentType string
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/storage/v1/object/resumes/cv.pdf" && r.Method == http.MethodPost {
			capturedUpsertHeader = r.Header.Get("x-upsert")
			capturedAuthHeader = r.Header.Get("Authorization")
			capturedContentType = r.Header.Get("Content-Type")
			capturedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	os.Setenv("SUPABASE_URL", server.URL)
	os.Setenv("SUPABASE_SERVICE_ROLE_KEY", "test-service-key")
	config.Load()
	defer func() {
		os.Unsetenv("SUPABASE_URL")
		os.Unsetenv("SUPABASE_SERVICE_ROLE_KEY")
		config.Load()
	}()

	testPDF := []byte("%PDF-1.4 mock pdf content")
	uri, err := UploadResumePDF(context.Background(), bytes.NewReader(testPDF), "application/pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if uri != "supabase://resumes/cv.pdf" {
		t.Errorf("expected uri supabase://resumes/cv.pdf, got %s", uri)
	}
	if capturedUpsertHeader != "true" {
		t.Errorf("expected x-upsert header to be 'true', got '%s'", capturedUpsertHeader)
	}
	if capturedAuthHeader != "Bearer test-service-key" {
		t.Errorf("expected Authorization header 'Bearer test-service-key', got '%s'", capturedAuthHeader)
	}
	if capturedContentType != "application/pdf" {
		t.Errorf("expected Content-Type 'application/pdf', got '%s'", capturedContentType)
	}
	if string(capturedBody) != string(testPDF) {
		t.Errorf("captured body does not match uploaded body")
	}
}

func TestDownloadResumePDF(t *testing.T) {
	mockData := []byte("%PDF-1.4 downloaded pdf content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/storage/v1/object/authenticated/resumes/cv.pdf" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write(mockData)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	os.Setenv("SUPABASE_URL", server.URL)
	os.Setenv("SUPABASE_SERVICE_ROLE_KEY", "test-service-key")
	config.Load()
	defer func() {
		os.Unsetenv("SUPABASE_URL")
		os.Unsetenv("SUPABASE_SERVICE_ROLE_KEY")
		config.Load()
	}()

	data, err := DownloadResumePDF(context.Background(), "supabase://resumes/cv.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(data) != string(mockData) {
		t.Errorf("expected '%s', got '%s'", string(mockData), string(data))
	}
}
