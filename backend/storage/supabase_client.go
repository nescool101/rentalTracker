package storage

import (
	"log"
	"os"

	supa "github.com/supabase-community/supabase-go"
)

// SupabaseClient is the client for Supabase
var SupabaseClient *supa.Client

// InitializeSupabaseClient initializes the Supabase client
func InitializeSupabaseClient() (*supa.Client, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		log.Fatal("❌ ERROR: SUPABASE_URL environment variable is required")
	}

	supabaseKey := os.Getenv("SUPABASE_KEY")
	if supabaseKey == "" {
		log.Fatal("❌ ERROR: SUPABASE_KEY environment variable is required")
	}

	client, err := supa.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		return nil, err
	}

	SupabaseClient = client
	log.Printf("✅ Supabase client initialized successfully")
	return client, nil
}

// GetSupabaseClient returns the Supabase client
func GetSupabaseClient() *supa.Client {
	if SupabaseClient == nil {
		client, err := InitializeSupabaseClient()
		if err != nil {
			panic(err)
		}
		return client
	}
	return SupabaseClient
}
