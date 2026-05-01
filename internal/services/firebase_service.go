package services

import (
	"context"
	"fmt"
	"log"

	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

var (
	FirebaseApp    *firebase.App
	FirebaseAuth   *auth.Client
	FirestoreClient *firestore.Client
)

// InitFirebase initializes the Firebase Admin SDK using Application Default Credentials (ADC).
func InitFirebase(ctx context.Context) error {
	var err error
	
	// Initialize Firebase App with ADC
	// This will look for GOOGLE_APPLICATION_CREDENTIALS environment variable
	// or use metadata service if running on GCP.
	FirebaseApp, err = firebase.NewApp(ctx, nil)
	if err != nil {
		return fmt.Errorf("error initializing firebase app: %v", err)
	}

	// Initialize Firebase Auth Client
	FirebaseAuth, err = FirebaseApp.Auth(ctx)
	if err != nil {
		return fmt.Errorf("error getting firebase auth client: %v", err)
	}

	// Initialize Firestore Client with specific database ID
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT environment variable not set")
	}

	FirestoreClient, err = firestore.NewClientWithDatabase(ctx, projectID, "portfolio")
	if err != nil {
		return fmt.Errorf("error getting firestore client: %v", err)
	}

	log.Println("✓ Firebase Admin SDK and Firestore (portfolio) initialized successfully")
	return nil
}
