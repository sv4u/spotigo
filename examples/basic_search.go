package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sv4u/spotigo"
)

// This example demonstrates searching for tracks without user authentication
// using the Client Credentials OAuth2 flow.
//
// Prerequisites:
//   - Set SPOTIPY_CLIENT_ID environment variable
//   - Set SPOTIPY_CLIENT_SECRET environment variable
//
// Usage:
//   go run examples/basic_search.go
func main() {
	// Get credentials from environment variables
	clientID := os.Getenv("SPOTIPY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIPY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("SPOTIPY_CLIENT_ID and SPOTIPY_CLIENT_SECRET must be set")
	}

	// Create authentication manager using Client Credentials flow
	// This flow doesn't require user authentication and is perfect for
	// accessing public Spotify data
	auth, err := spotigo.NewClientCredentials(clientID, clientSecret)
	if err != nil {
		log.Fatalf("Failed to create auth: %v", err)
	}

	// Create client
	client, err := spotigo.NewClient(auth)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Search for tracks matching "weezer" with limit of 20
	// This matches the Spotipy README example
	results, err := client.Search(ctx, "weezer", "track", &spotigo.SearchOptions{
		Limit: 20,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	// Display results
	fmt.Println("Search Results:")
	fmt.Println("==============")
	for i, track := range results.Tracks.Items {
		artistName := "Unknown"
		if len(track.Artists) > 0 {
			artistName = track.Artists[0].Name
		}
		fmt.Printf("%d. %s by %s\n", i+1, track.Name, artistName)
	}

	if results.Tracks.Total > 20 {
		fmt.Printf("\n... and %d more results\n", results.Tracks.Total-20)
	}
}
