package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	// Server
	Port string

	// Auth & Security
	AdminEmails            []string
	SupabaseURL            string
	SupabaseAnonKey        string
	SupabaseServiceRoleKey string
	SupabaseJWTSecret      string
	SupabaseDBURL          string

	// Turnstile
	TurnstileSiteKey   string
	TurnstileSecretKey string

	// AI
	NvidiaNIMAPI       string
	NvidiaNIMBaseURL   string
	NvidiaAPIKeyBackup string

	// Email
	EmailFrom    string
	ResendAPIKey string
}

var AppConfig Config

func Load() {
	AppConfig = Config{
		Port:                   getEnvOrDefault("PORT", "8080"),
		AdminEmails:            parseEmails(os.Getenv("ADMIN_EMAIL")),
		SupabaseURL:            os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:        os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseJWTSecret:      os.Getenv("SUPABASE_JWT_SECRET"),
		SupabaseDBURL:          os.Getenv("SUPABASE_DB_URL"),
		TurnstileSiteKey:       os.Getenv("TURNSTILE_SITE_KEY"),
		TurnstileSecretKey:     os.Getenv("TURNSTILE_SECRET_KEY"),
		NvidiaNIMAPI:           os.Getenv("NVIDIA_NIM_API"),
		NvidiaNIMBaseURL:       getEnvOrDefault("NVIDIA_NIM_BASE_URL", "https://integrate.api.nvidia.com/v1"),
		NvidiaAPIKeyBackup:     os.Getenv("NVIDIA_API_KEY_BACKUP"),
		EmailFrom:              os.Getenv("EMAIL_FROM"),
		ResendAPIKey:           os.Getenv("RESEND_API_KEY"),
	}

	if len(AppConfig.AdminEmails) == 0 {
		log.Println("CRITICAL SECURITY WARNING: ADMIN_EMAIL environment variable is not set. Access denied to all users.")
	}
}

func getEnvOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func parseEmails(raw string) []string {
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	var emails []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			emails = append(emails, trimmed)
		}
	}
	return emails
}
