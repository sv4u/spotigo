package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sv4u/spotigo"
	"github.com/sv4u/spotigo/tests"
)

func TestTrackEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tracks/6b2oQwSGFkzsMtQruIWm2p" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "6b2oQwSGFkzsMtQruIWm2p",
			"name": "Creep",
			"type": "track",
		})
	}))
	defer server.Close()

	auth := &tests.MockAuthManager{
		Token: &spotigo.TokenInfo{
			AccessToken: "test_token",
			TokenType:   "Bearer",
		},
	}

	client, err := spotigo.NewClient(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client.APIPrefix = server.URL + "/"

	ctx := context.Background()
	track, err := client.Track(ctx, "6b2oQwSGFkzsMtQruIWm2p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if track == nil {
		t.Fatal("expected track, got nil")
	}

	if track.ID != "6b2oQwSGFkzsMtQruIWm2p" {
		t.Errorf("expected ID '6b2oQwSGFkzsMtQruIWm2p', got %q", track.ID)
	}
}

func TestTracksEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tracks": []map[string]interface{}{
				{"id": "6b2oQwSGFkzsMtQruIWm2p", "name": "Creep"},
				{"id": "0Svkvt5I79wficMFgaqEQJ", "name": "El Scorcho"},
			},
		})
	}))
	defer server.Close()

	auth := &tests.MockAuthManager{
		Token: &spotigo.TokenInfo{
			AccessToken: "test_token",
			TokenType:   "Bearer",
		},
	}

	client, err := spotigo.NewClient(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client.APIPrefix = server.URL + "/"

	ctx := context.Background()
	result, err := client.Tracks(ctx, []string{"6b2oQwSGFkzsMtQruIWm2p", "0Svkvt5I79wficMFgaqEQJ"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected tracks response, got nil")
	}

	if len(result.Tracks) != 2 {
		t.Errorf("expected 2 tracks, got %d", len(result.Tracks))
	}
}

func TestArtistEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "3jOstUTkEu2JkjvRdBA5Gu",
			"name": "Weezer",
			"type": "artist",
		})
	}))
	defer server.Close()

	auth := &tests.MockAuthManager{
		Token: &spotigo.TokenInfo{
			AccessToken: "test_token",
			TokenType:   "Bearer",
		},
	}

	client, err := spotigo.NewClient(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client.APIPrefix = server.URL + "/"

	ctx := context.Background()
	artist, err := client.Artist(ctx, "3jOstUTkEu2JkjvRdBA5Gu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if artist == nil {
		t.Fatal("expected artist, got nil")
	}

	if artist.ID != "3jOstUTkEu2JkjvRdBA5Gu" {
		t.Errorf("expected ID '3jOstUTkEu2JkjvRdBA5Gu', got %q", artist.ID)
	}
}

func TestSearchEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "weezer" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("q"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tracks": map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": "6b2oQwSGFkzsMtQruIWm2p", "name": "Creep"},
				},
				"total": 1,
			},
		})
	}))
	defer server.Close()

	auth := &tests.MockAuthManager{
		Token: &spotigo.TokenInfo{
			AccessToken: "test_token",
			TokenType:   "Bearer",
		},
	}

	client, err := spotigo.NewClient(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client.APIPrefix = server.URL + "/"

	ctx := context.Background()
	result, err := client.Search(ctx, "weezer", "track", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected search result, got nil")
	}

	if result.Tracks == nil {
		t.Fatal("expected tracks in search result")
	}
}

func TestTracksMaxLimit(t *testing.T) {
	auth := &tests.MockAuthManager{
		Token: &spotigo.TokenInfo{
			AccessToken: "test_token",
			TokenType:   "Bearer",
		},
	}

	client, err := spotigo.NewClient(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	// Create 51 track IDs (exceeds max of 50)
	trackIDs := make([]string, 51)
	for i := 0; i < 51; i++ {
		trackIDs[i] = "6b2oQwSGFkzsMtQruIWm2p" // Reuse valid ID
	}

	_, err = client.Tracks(ctx, trackIDs)
	if err == nil {
		t.Fatal("expected error for exceeding max limit, got nil")
	}
}
