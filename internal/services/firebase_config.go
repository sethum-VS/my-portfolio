package services

import "os"

// FirebaseClientConfig holds the public Firebase client configuration values
// used by the frontend authentication page. These are injected from environment
// variables to avoid hardcoding them in templates.
type FirebaseClientConfig struct {
	APIKey            string `json:"apiKey"`
	AuthDomain        string `json:"authDomain"`
	ProjectID         string `json:"projectId"`
	StorageBucket     string `json:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId"`
	AppID             string `json:"appId"`
}

// GetFirebaseClientConfig reads the Firebase client config from environment
// variables. Falls back to reasonable defaults based on the project ID.
func GetFirebaseClientConfig() FirebaseClientConfig {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	return FirebaseClientConfig{
		APIKey:            envOrDefault("FIREBASE_API_KEY", "AIzaSyBHUsZvFm1e3VCQJjRuYyMo3-44dohxjcE"),
		AuthDomain:        envOrDefault("FIREBASE_AUTH_DOMAIN", projectID+".firebaseapp.com"),
		ProjectID:         projectID,
		StorageBucket:     envOrDefault("FIREBASE_STORAGE_BUCKET", projectID+".firebasestorage.app"),
		MessagingSenderID: envOrDefault("FIREBASE_MESSAGING_SENDER_ID", "1047596610069"),
		AppID:             envOrDefault("FIREBASE_APP_ID", "1:1047596610069:web:fb8aa48c22d07c18b2d41c"),
	}
}

func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
