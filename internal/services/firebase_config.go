package services

import (
	"log"
	"os"
	"sync"
)

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

var (
	firebaseClientConfig     FirebaseClientConfig
	firebaseClientConfigOnce sync.Once
)

// GetFirebaseClientConfig returns Firebase web client config from environment
// variables. Loads once; missing required vars terminate the process via log.Fatal.
func GetFirebaseClientConfig() FirebaseClientConfig {
	firebaseClientConfigOnce.Do(func() {
		firebaseClientConfig = loadFirebaseClientConfig()
	})
	return firebaseClientConfig
}

func loadFirebaseClientConfig() FirebaseClientConfig {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	apiKey := os.Getenv("FIREBASE_API_KEY")
	authDomain := os.Getenv("FIREBASE_AUTH_DOMAIN")
	storageBucket := os.Getenv("FIREBASE_STORAGE_BUCKET")
	messagingSenderID := os.Getenv("FIREBASE_MESSAGING_SENDER_ID")
	appID := os.Getenv("FIREBASE_APP_ID")

	if projectID == "" || apiKey == "" || authDomain == "" || storageBucket == "" || messagingSenderID == "" || appID == "" {
		log.Fatal("FATAL: Missing required Firebase environment variables")
	}

	return FirebaseClientConfig{
		APIKey:            apiKey,
		AuthDomain:        authDomain,
		ProjectID:         projectID,
		StorageBucket:     storageBucket,
		MessagingSenderID: messagingSenderID,
		AppID:             appID,
	}
}
