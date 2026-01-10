package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sv4u/spotigo"
)

// This example demonstrates accessing user-specific data using OAuth flow.
//
// Prerequisites:
//   - Set SPOTIPY_CLIENT_ID environment variable
//   - Set SPOTIPY_CLIENT_SECRET environment variable
//   - Set SPOTIPY_REDIRECT_URI environment variable (e.g., http://localhost:8080/callback)
//   - Add redirect URI to your Spotify app settings
//   - Set SPOTIPY_CLIENT_USERNAME environment variable (optional, for token caching)
//
// Usage:
//   go run examples/user_profile.go
func main() {
	clientID := os.Getenv("SPOTIPY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIPY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIPY_REDIRECT_URI")

	if clientID == "" || clientSecret == "" {
		log.Fatal("SPOTIPY_CLIENT_ID and SPOTIPY_CLIENT_SECRET must be set")
	}

	if redirectURI == "" {
		log.Fatal("SPOTIPY_REDIRECT_URI must be set")
	}

	// Create OAuth manager with required scopes
	// These scopes allow reading user profile and saved tracks
	auth, err := spotigo.NewSpotifyOAuth(
		clientID,
		clientSecret,
		redirectURI,
		"user-read-private user-read-email user-library-read",
	)
	if err != nil {
		log.Fatalf("Failed to create OAuth: %v", err)
	}

	ctx := context.Background()

	// Check if we have a cached token
	token, err := auth.GetCachedToken(ctx)
	if err != nil {
		log.Fatalf("Failed to get cached token: %v", err)
	}

	// If no cached token, get authorization
	if token == nil {
		fmt.Println("No cached token found. Opening browser for authorization...")
		code, err := auth.GetAuthorizationCode(ctx, true)
		if err != nil {
			log.Fatalf("Failed to get authorization code: %v", err)
		}

		// Exchange code for tokens
		if err := auth.ExchangeCode(ctx, code); err != nil {
			log.Fatalf("Failed to exchange code: %v", err)
		}
		fmt.Println("Authorization successful!")
	} else {
		fmt.Println("Using cached token.")
	}

	// Create client
	client, err := spotigo.NewClient(auth)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get current user profile
	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}

	fmt.Printf("\nLogged in as: %s (%s)\n", user.DisplayName, user.ID)
	if user.Email != nil {
		fmt.Printf("Email: %s\n", *user.Email)
	}
	fmt.Printf("Followers: %d\n", user.Followers.Total)
	fmt.Printf("Country: %s\n", user.Country)

	// Get saved tracks
	fmt.Println("\nYour Saved Tracks:")
	fmt.Println("==================")
	tracks, err := client.CurrentUserSavedTracks(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get saved tracks: %v", err)
	}

	// Display saved tracks in the format: "{artist_name} – {track_name}"
	// This matches the Spotipy README example format
	for _, item := range tracks.Items {
		artistName := "Unknown"
		if len(item.Track.Artists) > 0 {
			artistName = item.Track.Artists[0].Name
		}
		fmt.Printf("  %s – %s\n", artistName, item.Track.Name)
	}

	if tracks.Total > len(tracks.Items) {
		fmt.Printf("\n... and %d more saved tracks\n", tracks.Total-len(tracks.Items))
	}
}
