package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

// ============================================================================
// Artist Endpoints
// ============================================================================

// TestArtistsEndpoint tests the Artists endpoint (multiple artists)
func TestArtistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/artists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check query parameters
		ids := r.URL.Query().Get("ids")
		if ids == "" {
			t.Error("expected ids query parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"artists": []map[string]interface{}{
				{"id": "3jOstUTkEu2JkjvRdBA5Gu", "name": "Weezer", "type": "artist"},
				{"id": "1vCWHaC5f2uS3yhpwWbIA6", "name": "Avicii", "type": "artist"},
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
	result, err := client.Artists(ctx, []string{"3jOstUTkEu2JkjvRdBA5Gu", "1vCWHaC5f2uS3yhpwWbIA6"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected artists response, got nil")
	}

	if len(result.Artists) != 2 {
		t.Errorf("expected 2 artists, got %d", len(result.Artists))
	}

	if result.Artists[0].ID != "3jOstUTkEu2JkjvRdBA5Gu" {
		t.Errorf("expected first artist ID '3jOstUTkEu2JkjvRdBA5Gu', got %q", result.Artists[0].ID)
	}
}

// TestArtistsMaxLimit tests that Artists endpoint enforces max limit of 50
func TestArtistsMaxLimit(t *testing.T) {
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

	// Create 51 artist IDs (exceeds max of 50)
	artistIDs := make([]string, 51)
	for i := 0; i < 51; i++ {
		artistIDs[i] = "3jOstUTkEu2JkjvRdBA5Gu" // Reuse valid ID
	}

	_, err = client.Artists(ctx, artistIDs)
	if err == nil {
		t.Fatal("expected error for exceeding max limit, got nil")
	}

	if !strings.Contains(err.Error(), "maximum 50") {
		t.Errorf("expected error about maximum 50, got: %v", err)
	}
}

// TestArtistRelatedArtistsEndpoint tests the ArtistRelatedArtists endpoint
func TestArtistRelatedArtistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/artists/3jOstUTkEu2JkjvRdBA5Gu/related-artists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"artists": []map[string]interface{}{
				{"id": "1vCWHaC5f2uS3yhpwWbIA6", "name": "Related Artist 1", "type": "artist"},
				{"id": "4Z8W4fKeB5YxbusRsdQVPb", "name": "Related Artist 2", "type": "artist"},
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
	result, err := client.ArtistRelatedArtists(ctx, "3jOstUTkEu2JkjvRdBA5Gu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected artists response, got nil")
	}

	if len(result.Artists) != 2 {
		t.Errorf("expected 2 related artists, got %d", len(result.Artists))
	}
}

// ============================================================================
// Album Endpoints
// ============================================================================

// TestAlbumEndpoint tests the Album endpoint (single album)
func TestAlbumEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/albums/04xe676vyiTeYNXw15o9jT" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "04xe676vyiTeYNXw15o9jT",
			"name": "Pinkerton",
			"type": "album",
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
	album, err := client.Album(ctx, "04xe676vyiTeYNXw15o9jT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if album == nil {
		t.Fatal("expected album, got nil")
	}

	if album.ID != "04xe676vyiTeYNXw15o9jT" {
		t.Errorf("expected ID '04xe676vyiTeYNXw15o9jT', got %q", album.ID)
	}
}

// TestAlbumEndpointWithMarket tests the Album endpoint with market parameter
func TestAlbumEndpointWithMarket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "04xe676vyiTeYNXw15o9jT",
			"name": "Pinkerton",
			"type": "album",
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
	album, err := client.Album(ctx, "04xe676vyiTeYNXw15o9jT", "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if album == nil {
		t.Fatal("expected album, got nil")
	}
}

// TestAlbumsEndpoint tests the Albums endpoint (multiple albums)
func TestAlbumsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/albums" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check query parameters
		ids := r.URL.Query().Get("ids")
		if ids == "" {
			t.Error("expected ids query parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"albums": []map[string]interface{}{
				{"id": "04xe676vyiTeYNXw15o9jT", "name": "Pinkerton", "type": "album"},
				{"id": "1ATL5GLyefJaxhQzSPVrGW", "name": "Blue Album", "type": "album"},
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
	result, err := client.Albums(ctx, []string{"04xe676vyiTeYNXw15o9jT", "1ATL5GLyefJaxhQzSPVrGW"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected albums response, got nil")
	}

	if len(result.Albums) != 2 {
		t.Errorf("expected 2 albums, got %d", len(result.Albums))
	}
}

// TestAlbumsMaxLimit tests that Albums endpoint enforces max limit of 20
func TestAlbumsMaxLimit(t *testing.T) {
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

	// Create 21 album IDs (exceeds max of 20)
	albumIDs := make([]string, 21)
	for i := 0; i < 21; i++ {
		albumIDs[i] = "04xe676vyiTeYNXw15o9jT" // Reuse valid ID
	}

	_, err = client.Albums(ctx, albumIDs)
	if err == nil {
		t.Fatal("expected error for exceeding max limit, got nil")
	}

	if !strings.Contains(err.Error(), "maximum 20") {
		t.Errorf("expected error about maximum 20, got: %v", err)
	}
}

// TestAlbumTracksEndpoint tests the AlbumTracks endpoint
func TestAlbumTracksEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/albums/04xe676vyiTeYNXw15o9jT/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "6b2oQwSGFkzsMtQruIWm2p", "name": "Track 1", "type": "track"},
				{"id": "0Svkvt5I79wficMFgaqEQJ", "name": "Track 2", "type": "track"},
			},
			"total":  2,
			"limit":  20,
			"offset": 0,
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
	result, err := client.AlbumTracks(ctx, "04xe676vyiTeYNXw15o9jT", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected tracks response, got nil")
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 tracks, got %d", len(result.Items))
	}
}

// TestAlbumTracksEndpointWithOptions tests the AlbumTracks endpoint with options
func TestAlbumTracksEndpointWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		offset := r.URL.Query().Get("offset")
		if offset != "5" {
			t.Errorf("expected offset=5, got %q", offset)
		}

		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  10,
			"offset": 5,
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
	opts := &spotigo.AlbumTracksOptions{
		Market: "US",
		Limit:  10,
		Offset: 5,
	}
	result, err := client.AlbumTracks(ctx, "04xe676vyiTeYNXw15o9jT", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected tracks response, got nil")
	}
}

// ============================================================================
// Playlist Endpoints
// ============================================================================

// TestPlaylistEndpoint tests the Playlist endpoint (single playlist)
func TestPlaylistEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "2oCEWyyAPbZp9xhVSxZavx",
			"name": "Test Playlist",
			"type": "playlist",
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
	playlist, err := client.Playlist(ctx, "2oCEWyyAPbZp9xhVSxZavx", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if playlist == nil {
		t.Fatal("expected playlist, got nil")
	}

	if playlist.ID != "2oCEWyyAPbZp9xhVSxZavx" {
		t.Errorf("expected ID '2oCEWyyAPbZp9xhVSxZavx', got %q", playlist.ID)
	}
}

// TestPlaylistEndpointWithOptions tests the Playlist endpoint with options
func TestPlaylistEndpointWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := r.URL.Query().Get("fields")
		if fields != "name,description" {
			t.Errorf("expected fields=name,description, got %q", fields)
		}

		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "2oCEWyyAPbZp9xhVSxZavx",
			"name": "Test Playlist",
			"type": "playlist",
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
	opts := &spotigo.PlaylistOptions{
		Fields: "name,description",
		Market: "US",
	}
	playlist, err := client.Playlist(ctx, "2oCEWyyAPbZp9xhVSxZavx", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if playlist == nil {
		t.Fatal("expected playlist, got nil")
	}
}

// TestPlaylistTracksEndpoint tests the PlaylistTracks endpoint
func TestPlaylistTracksEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"track": map[string]interface{}{"id": "6b2oQwSGFkzsMtQruIWm2p", "name": "Track 1"}},
				{"track": map[string]interface{}{"id": "0Svkvt5I79wficMFgaqEQJ", "name": "Track 2"}},
			},
			"total":  2,
			"limit":  100,
			"offset": 0,
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
	result, err := client.PlaylistTracks(ctx, "2oCEWyyAPbZp9xhVSxZavx", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected tracks response, got nil")
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 tracks, got %d", len(result.Items))
	}
}

// TestCurrentUserPlaylistsEndpoint tests the CurrentUserPlaylists endpoint
func TestCurrentUserPlaylistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/playlists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "2oCEWyyAPbZp9xhVSxZavx", "name": "Playlist 1", "type": "playlist"},
				{"id": "37i9dQZF1DXcBWIGoYBM5M", "name": "Playlist 2", "type": "playlist"},
			},
			"total":  2,
			"limit":  20,
			"offset": 0,
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
	result, err := client.CurrentUserPlaylists(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected playlists response, got nil")
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 playlists, got %d", len(result.Items))
	}
}

// TestUserPlaylistCreateEndpoint tests the UserPlaylistCreate endpoint
func TestUserPlaylistCreateEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/users/testuser/playlists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Parse request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["name"] != "New Playlist" {
			t.Errorf("expected name 'New Playlist', got %v", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "new_playlist_id",
			"name": "New Playlist",
			"type": "playlist",
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
	opts := &spotigo.CreatePlaylistOptions{
		Name: "New Playlist",
	}
	playlist, err := client.UserPlaylistCreate(ctx, "testuser", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if playlist == nil {
		t.Fatal("expected playlist, got nil")
	}

	if playlist.ID != "new_playlist_id" {
		t.Errorf("expected ID 'new_playlist_id', got %q", playlist.ID)
	}
}

// TestUserPlaylistCreateEndpointValidation tests validation for UserPlaylistCreate
func TestUserPlaylistCreateEndpointValidation(t *testing.T) {
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

	// Test: nil options
	_, err = client.UserPlaylistCreate(ctx, "testuser", nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}

	// Test: empty name
	opts := &spotigo.CreatePlaylistOptions{
		Name: "",
	}
	_, err = client.UserPlaylistCreate(ctx, "testuser", opts)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

// TestPlaylistCoverImageEndpoint tests the PlaylistCoverImage endpoint
func TestPlaylistCoverImageEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx/images" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"url": "https://i.scdn.co/image/abc123", "height": 640, "width": 640},
			{"url": "https://i.scdn.co/image/def456", "height": 300, "width": 300},
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
	images, err := client.PlaylistCoverImage(ctx, "2oCEWyyAPbZp9xhVSxZavx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if images == nil {
		t.Fatal("expected images, got nil")
	}

	if len(images) != 2 {
		t.Errorf("expected 2 images, got %d", len(images))
	}
}

// TestPlaylistReplaceItemsEndpoint tests the PlaylistReplaceItems endpoint
func TestPlaylistReplaceItemsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Parse request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		uris, ok := body["uris"].([]interface{})
		if !ok {
			t.Error("expected uris array in request body")
		}

		if len(uris) == 0 {
			t.Error("expected at least one URI")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"snapshot_id": "new_snapshot_id",
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
	items := []string{"spotify:track:6b2oQwSGFkzsMtQruIWm2p", "spotify:track:0Svkvt5I79wficMFgaqEQJ"}
	result, err := client.PlaylistReplaceItems(ctx, "2oCEWyyAPbZp9xhVSxZavx", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected snapshot ID, got nil")
	}

	if result.SnapshotID != "new_snapshot_id" {
		t.Errorf("expected snapshot_id 'new_snapshot_id', got %q", result.SnapshotID)
	}
}

// TestPlaylistReorderItemsEndpoint tests the PlaylistReorderItems endpoint
func TestPlaylistReorderItemsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		// Parse request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["range_start"] != float64(0) {
			t.Errorf("expected range_start=0, got %v", body["range_start"])
		}

		if body["insert_before"] != float64(2) {
			t.Errorf("expected insert_before=2, got %v", body["insert_before"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"snapshot_id": "reordered_snapshot_id",
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
	opts := &spotigo.ReorderItemsOptions{
		RangeStart:  0,
		InsertBefore: 2,
	}
	result, err := client.PlaylistReorderItems(ctx, "2oCEWyyAPbZp9xhVSxZavx", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected snapshot ID, got nil")
	}
}

// TestPlaylistReorderItemsValidation tests validation for PlaylistReorderItems
func TestPlaylistReorderItemsValidation(t *testing.T) {
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

	// Test: nil options
	_, err = client.PlaylistReorderItems(ctx, "2oCEWyyAPbZp9xhVSxZavx", nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}
}

// TestPlaylistChangeDetailsEndpoint tests the PlaylistChangeDetails endpoint
func TestPlaylistChangeDetailsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Parse request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		name, ok := body["name"].(string)
		if !ok || name != "Updated Playlist Name" {
			t.Errorf("expected name 'Updated Playlist Name', got %v", body["name"])
		}

		w.WriteHeader(http.StatusOK)
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
	name := "Updated Playlist Name"
	public := true
	opts := &spotigo.ChangePlaylistDetailsOptions{
		Name:   &name,
		Public: &public,
	}
	err = client.PlaylistChangeDetails(ctx, "2oCEWyyAPbZp9xhVSxZavx", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPlaylistChangeDetailsValidation tests validation for PlaylistChangeDetails
func TestPlaylistChangeDetailsValidation(t *testing.T) {
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

	// Test: nil options
	err = client.PlaylistChangeDetails(ctx, "2oCEWyyAPbZp9xhVSxZavx", nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}
}

// TestPlaylistUploadCoverImageEndpoint tests the PlaylistUploadCoverImage endpoint
func TestPlaylistUploadCoverImageEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx/images" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "image/jpeg" {
			t.Errorf("expected Content-Type 'image/jpeg', got %q", contentType)
		}

		w.WriteHeader(http.StatusAccepted)
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
	// Create valid JPEG image data (JPEG magic bytes: FF D8)
	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	err = client.PlaylistUploadCoverImage(ctx, "2oCEWyyAPbZp9xhVSxZavx", imageData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPlaylistUploadCoverImageValidation tests validation for PlaylistUploadCoverImage
func TestPlaylistUploadCoverImageValidation(t *testing.T) {
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

	// Test: Image too large (max 256KB)
	largeImage := make([]byte, 257*1024) // 257KB
	largeImage[0] = 0xFF
	largeImage[1] = 0xD8
	err = client.PlaylistUploadCoverImage(ctx, "2oCEWyyAPbZp9xhVSxZavx", largeImage)
	if err == nil {
		t.Fatal("expected error for image too large, got nil")
	}

	if !strings.Contains(err.Error(), "256KB") {
		t.Errorf("expected error about 256KB limit, got: %v", err)
	}

	// Test: Invalid JPEG format
	invalidImage := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	err = client.PlaylistUploadCoverImage(ctx, "2oCEWyyAPbZp9xhVSxZavx", invalidImage)
	if err == nil {
		t.Fatal("expected error for invalid JPEG format, got nil")
	}

	if !strings.Contains(err.Error(), "JPEG") {
		t.Errorf("expected error about JPEG format, got: %v", err)
	}
}

// ============================================================================
// User Endpoints
// ============================================================================

// TestCurrentUserEndpoint tests the CurrentUser endpoint
func TestCurrentUserEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "test_user_id",
			"display_name": "Test User",
			"type":        "user",
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
	user, err := client.CurrentUser(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user == nil {
		t.Fatal("expected user, got nil")
	}

	if user.ID != "test_user_id" {
		t.Errorf("expected ID 'test_user_id', got %q", user.ID)
	}
}

// TestUserEndpoint tests the User endpoint (public user profile)
func TestUserEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/testuser" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "testuser",
			"display_name": "Test User",
			"type":        "user",
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
	user, err := client.User(ctx, "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user == nil {
		t.Fatal("expected user, got nil")
	}

	if user.ID != "testuser" {
		t.Errorf("expected ID 'testuser', got %q", user.ID)
	}
}

// TestCurrentUserSavedTracksEndpoint tests the CurrentUserSavedTracks endpoint
func TestCurrentUserSavedTracksEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"track": map[string]interface{}{"id": "6b2oQwSGFkzsMtQruIWm2p", "name": "Track 1"}},
				{"track": map[string]interface{}{"id": "0Svkvt5I79wficMFgaqEQJ", "name": "Track 2"}},
			},
			"total":  2,
			"limit":  20,
			"offset": 0,
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
	result, err := client.CurrentUserSavedTracks(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved tracks response, got nil")
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 saved tracks, got %d", len(result.Items))
	}
}

// TestCurrentUserSavedTracksAddEndpoint tests the CurrentUserSavedTracksAdd endpoint
func TestCurrentUserSavedTracksAddEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/me/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Parse request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		ids, ok := body["ids"].([]interface{})
		if !ok {
			t.Error("expected ids array in request body")
		}

		if len(ids) == 0 {
			t.Error("expected at least one ID")
		}

		w.WriteHeader(http.StatusOK)
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
	trackIDs := []string{"6b2oQwSGFkzsMtQruIWm2p", "0Svkvt5I79wficMFgaqEQJ"}
	err = client.CurrentUserSavedTracksAdd(ctx, trackIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSavedTracksDeleteEndpoint tests the CurrentUserSavedTracksDelete endpoint
func TestCurrentUserSavedTracksDeleteEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/me/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Parse request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		ids, ok := body["ids"].([]interface{})
		if !ok {
			t.Error("expected ids array in request body")
		}

		if len(ids) == 0 {
			t.Error("expected at least one ID")
		}

		w.WriteHeader(http.StatusOK)
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
	trackIDs := []string{"6b2oQwSGFkzsMtQruIWm2p"}
	err = client.CurrentUserSavedTracksDelete(ctx, trackIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSavedTracksContainsEndpoint tests the CurrentUserSavedTracksContains endpoint
func TestCurrentUserSavedTracksContainsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/tracks/contains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check query parameters
		ids := r.URL.Query().Get("ids")
		if ids == "" {
			t.Error("expected ids query parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]bool{true, false})
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
	trackIDs := []string{"6b2oQwSGFkzsMtQruIWm2p", "0Svkvt5I79wficMFgaqEQJ"}
	result, err := client.CurrentUserSavedTracksContains(ctx, trackIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected contains result, got nil")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}

	if !result[0] {
		t.Error("expected first track to be saved")
	}

	if result[1] {
		t.Error("expected second track to not be saved")
	}
}

// TestCurrentUserSavedAlbumsEndpoint tests the CurrentUserSavedAlbums endpoint
func TestCurrentUserSavedAlbumsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/albums" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"album": map[string]interface{}{"id": "04xe676vyiTeYNXw15o9jT", "name": "Album 1"}},
			},
			"total":  1,
			"limit":  20,
			"offset": 0,
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
	result, err := client.CurrentUserSavedAlbums(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved albums response, got nil")
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 saved album, got %d", len(result.Items))
	}
}

// TestCurrentUserSavedAlbumsAddEndpoint tests the CurrentUserSavedAlbumsAdd endpoint
func TestCurrentUserSavedAlbumsAddEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
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
	albumIDs := []string{"04xe676vyiTeYNXw15o9jT"}
	err = client.CurrentUserSavedAlbumsAdd(ctx, albumIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSavedAlbumsDeleteEndpoint tests the CurrentUserSavedAlbumsDelete endpoint
func TestCurrentUserSavedAlbumsDeleteEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
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
	albumIDs := []string{"04xe676vyiTeYNXw15o9jT"}
	err = client.CurrentUserSavedAlbumsDelete(ctx, albumIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSavedAlbumsContainsEndpoint tests the CurrentUserSavedAlbumsContains endpoint
func TestCurrentUserSavedAlbumsContainsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/albums/contains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]bool{true})
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
	albumIDs := []string{"04xe676vyiTeYNXw15o9jT"}
	result, err := client.CurrentUserSavedAlbumsContains(ctx, albumIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected contains result, got nil")
	}

	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

// TestCurrentUserSavedEpisodesEndpoint tests the CurrentUserSavedEpisodes endpoint
func TestCurrentUserSavedEpisodesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/episodes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  20,
			"offset": 0,
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
	result, err := client.CurrentUserSavedEpisodes(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved episodes response, got nil")
	}
}

// TestCurrentUserSavedShowsEndpoint tests the CurrentUserSavedShows endpoint
func TestCurrentUserSavedShowsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/shows" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  20,
			"offset": 0,
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
	result, err := client.CurrentUserSavedShows(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved shows response, got nil")
	}
}

// TestCurrentUserSavedEpisodesAddEndpoint tests the CurrentUserSavedEpisodesAdd endpoint
func TestCurrentUserSavedEpisodesAddEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
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
	episodeIDs := []string{"4rOoJ6Egrf8K2IrywzwOMk"}
	err = client.CurrentUserSavedEpisodesAdd(ctx, episodeIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSavedEpisodesDeleteEndpoint tests the CurrentUserSavedEpisodesDelete endpoint
func TestCurrentUserSavedEpisodesDeleteEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
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
	episodeIDs := []string{"4rOoJ6Egrf8K2IrywzwOMk"}
	err = client.CurrentUserSavedEpisodesDelete(ctx, episodeIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSavedEpisodesContainsEndpoint tests the CurrentUserSavedEpisodesContains endpoint
func TestCurrentUserSavedEpisodesContainsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/episodes/contains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]bool{true, false})
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
	episodeIDs := []string{"4rOoJ6Egrf8K2IrywzwOMk", "5AvwZVawapvyqJNqo3S7yF"}
	result, err := client.CurrentUserSavedEpisodesContains(ctx, episodeIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected contains result, got nil")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

// TestCurrentUserSavedShowsAddEndpoint tests the CurrentUserSavedShowsAdd endpoint
func TestCurrentUserSavedShowsAddEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
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
	showIDs := []string{"4rOoJ6Egrf8K2IrywzwOMk"}
	err = client.CurrentUserSavedShowsAdd(ctx, showIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSavedShowsDeleteEndpoint tests the CurrentUserSavedShowsDelete endpoint
func TestCurrentUserSavedShowsDeleteEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
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
	showIDs := []string{"4rOoJ6Egrf8K2IrywzwOMk"}
	err = client.CurrentUserSavedShowsDelete(ctx, showIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSavedShowsContainsEndpoint tests the CurrentUserSavedShowsContains endpoint
func TestCurrentUserSavedShowsContainsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/shows/contains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]bool{true})
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
	showIDs := []string{"4rOoJ6Egrf8K2IrywzwOMk"}
	result, err := client.CurrentUserSavedShowsContains(ctx, showIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected contains result, got nil")
	}

	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

// ============================================================================
// Browse Endpoints
// ============================================================================

// TestBrowseCategoriesEndpoint tests the BrowseCategories endpoint
func TestBrowseCategoriesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/browse/categories" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"categories": map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": "pop", "name": "Pop"},
					{"id": "rock", "name": "Rock"},
				},
				"total":  2,
				"limit":  20,
				"offset": 0,
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
	result, err := client.BrowseCategories(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected categories response, got nil")
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 categories, got %d", len(result.Items))
	}
}

// TestBrowseCategoryEndpoint tests the BrowseCategory endpoint
func TestBrowseCategoryEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/browse/categories/pop" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "pop",
			"name": "Pop",
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
	category, err := client.BrowseCategory(ctx, "pop", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if category == nil {
		t.Fatal("expected category, got nil")
	}

	if category.ID != "pop" {
		t.Errorf("expected ID 'pop', got %q", category.ID)
	}
}

// TestBrowseFeaturedPlaylistsEndpoint tests the BrowseFeaturedPlaylists endpoint
func TestBrowseFeaturedPlaylistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/browse/featured-playlists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Featured Playlists",
			"playlists": map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": "2oCEWyyAPbZp9xhVSxZavx", "name": "Featured Playlist 1"},
				},
				"total":  1,
				"limit":  20,
				"offset": 0,
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
	result, err := client.BrowseFeaturedPlaylists(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected featured playlists response, got nil")
	}

	if result.Message != "Featured Playlists" {
		t.Errorf("expected message 'Featured Playlists', got %q", result.Message)
	}

	if len(result.Playlists.Items) != 1 {
		t.Errorf("expected 1 playlist, got %d", len(result.Playlists.Items))
	}
}

// TestBrowseNewReleasesEndpoint tests the BrowseNewReleases endpoint
func TestBrowseNewReleasesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/browse/new-releases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"albums": map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": "04xe676vyiTeYNXw15o9jT", "name": "New Album 1"},
				},
				"total":  1,
				"limit":  20,
				"offset": 0,
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
	result, err := client.BrowseNewReleases(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected new releases response, got nil")
	}

	if len(result.Albums.Items) != 1 {
		t.Errorf("expected 1 album, got %d", len(result.Albums.Items))
	}
}

// TestBrowseCategoryPlaylistsEndpoint tests the BrowseCategoryPlaylists endpoint
func TestBrowseCategoryPlaylistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/browse/categories/pop/playlists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"playlists": map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": "2oCEWyyAPbZp9xhVSxZavx", "name": "Pop Playlist 1"},
				},
				"total":  1,
				"limit":  20,
				"offset": 0,
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
	result, err := client.BrowseCategoryPlaylists(ctx, "pop", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected category playlists response, got nil")
	}

	if len(result.Playlists.Items) != 1 {
		t.Errorf("expected 1 playlist, got %d", len(result.Playlists.Items))
	}
}

// TestRecommendationsEndpoint tests the Recommendations endpoint
func TestRecommendationsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/recommendations" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		seedArtists := r.URL.Query().Get("seed_artists")
		if seedArtists == "" {
			t.Error("expected seed_artists query parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tracks": []map[string]interface{}{
				{"id": "6b2oQwSGFkzsMtQruIWm2p", "name": "Recommended Track 1"},
				{"id": "0Svkvt5I79wficMFgaqEQJ", "name": "Recommended Track 2"},
			},
			"seeds": []map[string]interface{}{
				{"id": "3jOstUTkEu2JkjvRdBA5Gu", "type": "artist"},
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
	opts := &spotigo.RecommendationsOptions{
		SeedArtists: []string{"3jOstUTkEu2JkjvRdBA5Gu"},
		Limit:       20,
	}
	result, err := client.Recommendations(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected recommendations response, got nil")
	}

	if len(result.Tracks) != 2 {
		t.Errorf("expected 2 recommended tracks, got %d", len(result.Tracks))
	}
}

// TestRecommendationsValidation tests validation for Recommendations endpoint
func TestRecommendationsValidation(t *testing.T) {
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

	_, err = client.Recommendations(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}

	opts := &spotigo.RecommendationsOptions{
		Limit: 20,
	}
	_, err = client.Recommendations(ctx, opts)
	if err == nil {
		t.Fatal("expected error for no seeds, got nil")
	}

	opts2 := &spotigo.RecommendationsOptions{
		SeedArtists: []string{"1", "2", "3", "4", "5", "6"},
		Limit:       20,
	}
	_, err = client.Recommendations(ctx, opts2)
	if err == nil {
		t.Fatal("expected error for too many seed artists, got nil")
	}
}

// ============================================================================
// Following Endpoints
// ============================================================================

// TestCurrentUserFollowedArtistsEndpoint tests the CurrentUserFollowedArtists endpoint
func TestCurrentUserFollowedArtistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/following" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		typeParam := r.URL.Query().Get("type")
		if typeParam != "artist" {
			t.Errorf("expected type=artist, got %q", typeParam)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"artists": map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": "3jOstUTkEu2JkjvRdBA5Gu", "name": "Weezer", "type": "artist"},
				},
				"total":  1,
				"limit":  20,
				"cursors": map[string]interface{}{
					"after": "cursor_after",
				},
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
	result, err := client.CurrentUserFollowedArtists(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected followed artists response, got nil")
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 artist, got %d", len(result.Items))
	}
}

// TestCurrentUserFollowingArtistsEndpoint tests the CurrentUserFollowingArtists endpoint
func TestCurrentUserFollowingArtistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/following/contains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		typeParam := r.URL.Query().Get("type")
		if typeParam != "artist" {
			t.Errorf("expected type=artist, got %q", typeParam)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]bool{true, false})
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
	artistIDs := []string{"3jOstUTkEu2JkjvRdBA5Gu", "1vCWHaC5f2uS3yhpwWbIA6"}
	result, err := client.CurrentUserFollowingArtists(ctx, artistIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected following result, got nil")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

// TestCurrentUserFollowingUsersEndpoint tests the CurrentUserFollowingUsers endpoint
func TestCurrentUserFollowingUsersEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/following/contains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		typeParam := r.URL.Query().Get("type")
		if typeParam != "user" {
			t.Errorf("expected type=user, got %q", typeParam)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]bool{true})
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
	userIDs := []string{"testuser"}
	result, err := client.CurrentUserFollowingUsers(ctx, userIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected following result, got nil")
	}

	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

// TestUserFollowArtistsEndpoint tests the UserFollowArtists endpoint
func TestUserFollowArtistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["type"] != "artist" {
			t.Errorf("expected type=artist, got %v", body["type"])
		}

		w.WriteHeader(http.StatusOK)
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
	artistIDs := []string{"3jOstUTkEu2JkjvRdBA5Gu"}
	err = client.UserFollowArtists(ctx, artistIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestUserFollowUsersEndpoint tests the UserFollowUsers endpoint
func TestUserFollowUsersEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["type"] != "user" {
			t.Errorf("expected type=user, got %v", body["type"])
		}

		w.WriteHeader(http.StatusOK)
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
	userIDs := []string{"testuser"}
	err = client.UserFollowUsers(ctx, userIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserFollowPlaylistEndpoint tests the CurrentUserFollowPlaylist endpoint
func TestCurrentUserFollowPlaylistEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx/followers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserFollowPlaylist(ctx, "2oCEWyyAPbZp9xhVSxZavx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserUnfollowPlaylistEndpoint tests the CurrentUserUnfollowPlaylist endpoint
func TestCurrentUserUnfollowPlaylistEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx/followers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserUnfollowPlaylist(ctx, "2oCEWyyAPbZp9xhVSxZavx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPlaylistIsFollowingEndpoint tests the PlaylistIsFollowing endpoint
func TestPlaylistIsFollowingEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playlists/2oCEWyyAPbZp9xhVSxZavx/followers/contains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]bool{true, false})
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
	userIDs := []string{"user1", "user2"}
	result, err := client.PlaylistIsFollowing(ctx, "2oCEWyyAPbZp9xhVSxZavx", userIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected following result, got nil")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

// ============================================================================
// Audio Features Endpoints
// ============================================================================

// TestAudioFeaturesEndpoint tests the AudioFeatures endpoint
func TestAudioFeaturesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio-features/6b2oQwSGFkzsMtQruIWm2p" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           "6b2oQwSGFkzsMtQruIWm2p",
			"danceability": 0.5,
			"energy":       0.7,
			"valence":      0.6,
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
	result, err := client.AudioFeatures(ctx, "6b2oQwSGFkzsMtQruIWm2p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected audio features, got nil")
	}

	if result.ID != "6b2oQwSGFkzsMtQruIWm2p" {
		t.Errorf("expected ID '6b2oQwSGFkzsMtQruIWm2p', got %q", result.ID)
	}
}

// TestAudioFeaturesMultipleEndpoint tests the AudioFeaturesMultiple endpoint
func TestAudioFeaturesMultipleEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio-features" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		ids := r.URL.Query().Get("ids")
		if ids == "" {
			t.Error("expected ids query parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"audio_features": []map[string]interface{}{
				{"id": "6b2oQwSGFkzsMtQruIWm2p", "danceability": 0.5},
				{"id": "0Svkvt5I79wficMFgaqEQJ", "danceability": 0.6},
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
	trackIDs := []string{"6b2oQwSGFkzsMtQruIWm2p", "0Svkvt5I79wficMFgaqEQJ"}
	result, err := client.AudioFeaturesMultiple(ctx, trackIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected audio features, got nil")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 audio features, got %d", len(result))
	}
}

// TestAudioAnalysisEndpoint tests the AudioAnalysis endpoint
func TestAudioAnalysisEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio-analysis/6b2oQwSGFkzsMtQruIWm2p" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"track": map[string]interface{}{
				"duration": 240.0,
				"tempo":    120.0,
			},
			"bars":    []map[string]interface{}{},
			"beats":   []map[string]interface{}{},
			"sections": []map[string]interface{}{},
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
	result, err := client.AudioAnalysis(ctx, "6b2oQwSGFkzsMtQruIWm2p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected audio analysis, got nil")
	}
}

// ============================================================================
// Show/Episode Endpoints
// ============================================================================

// TestShowEndpoint tests the Show endpoint
func TestShowEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shows/4rOoJ6Egrf8K2IrywzwOMk" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "4rOoJ6Egrf8K2IrywzwOMk",
			"name": "Test Show",
			"type": "show",
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
	show, err := client.Show(ctx, "4rOoJ6Egrf8K2IrywzwOMk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if show == nil {
		t.Fatal("expected show, got nil")
	}

	if show.ID != "4rOoJ6Egrf8K2IrywzwOMk" {
		t.Errorf("expected ID '4rOoJ6Egrf8K2IrywzwOMk', got %q", show.ID)
	}
}

// TestShowsEndpoint tests the Shows endpoint
func TestShowsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shows" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"shows": []map[string]interface{}{
				{"id": "4rOoJ6Egrf8K2IrywzwOMk", "name": "Show 1", "type": "show"},
				{"id": "5AvwZVawapvyqJNqo3S7yF", "name": "Show 2", "type": "show"},
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
	result, err := client.Shows(ctx, []string{"4rOoJ6Egrf8K2IrywzwOMk", "5AvwZVawapvyqJNqo3S7yF"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected shows response, got nil")
	}

	if len(result.Shows) != 2 {
		t.Errorf("expected 2 shows, got %d", len(result.Shows))
	}
}

// TestShowEpisodesEndpoint tests the ShowEpisodes endpoint
func TestShowEpisodesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shows/4rOoJ6Egrf8K2IrywzwOMk/episodes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "5AvwZVawapvyqJNqo3S7yF", "name": "Episode 1", "type": "episode"},
			},
			"total":  1,
			"limit":  20,
			"offset": 0,
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
	result, err := client.ShowEpisodes(ctx, "4rOoJ6Egrf8K2IrywzwOMk", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected episodes response, got nil")
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 episode, got %d", len(result.Items))
	}
}

// TestEpisodeEndpoint tests the Episode endpoint
func TestEpisodeEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/episodes/5AvwZVawapvyqJNqo3S7yF" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "5AvwZVawapvyqJNqo3S7yF",
			"name": "Test Episode",
			"type": "episode",
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
	episode, err := client.Episode(ctx, "5AvwZVawapvyqJNqo3S7yF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if episode == nil {
		t.Fatal("expected episode, got nil")
	}

	if episode.ID != "5AvwZVawapvyqJNqo3S7yF" {
		t.Errorf("expected ID '5AvwZVawapvyqJNqo3S7yF', got %q", episode.ID)
	}
}

// TestEpisodesEndpoint tests the Episodes endpoint
func TestEpisodesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/episodes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"episodes": []map[string]interface{}{
				{"id": "5AvwZVawapvyqJNqo3S7yF", "name": "Episode 1", "type": "episode"},
				{"id": "6BwZVawapvyqJNqo3S7yG", "name": "Episode 2", "type": "episode"},
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
	result, err := client.Episodes(ctx, []string{"5AvwZVawapvyqJNqo3S7yF", "6BwZVawapvyqJNqo3S7yG"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected episodes response, got nil")
	}

	if len(result.Episodes) != 2 {
		t.Errorf("expected 2 episodes, got %d", len(result.Episodes))
	}
}

// ============================================================================
// Playback Endpoints
// ============================================================================

// TestCurrentUserPlayingTrackEndpoint tests the CurrentUserPlayingTrack endpoint
func TestCurrentUserPlayingTrackEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/player/currently-playing" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"is_playing": true,
			"item": map[string]interface{}{
				"id":   "6b2oQwSGFkzsMtQruIWm2p",
				"name": "Creep",
				"type": "track",
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
	result, err := client.CurrentUserPlayingTrack(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected currently playing, got nil")
	}
}

// TestCurrentUserPlaybackStateEndpoint tests the CurrentUserPlaybackState endpoint
func TestCurrentUserPlaybackStateEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/player" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"is_playing": false,
			"device": map[string]interface{}{
				"id":   "device_id",
				"name": "Test Device",
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
	result, err := client.CurrentUserPlaybackState(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected playback state, got nil")
	}
}

// TestCurrentUserDevicesEndpoint tests the CurrentUserDevices endpoint
func TestCurrentUserDevicesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/player/devices" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"devices": []map[string]interface{}{
				{"id": "device1", "name": "Device 1", "type": "Computer"},
				{"id": "device2", "name": "Device 2", "type": "Smartphone"},
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
	result, err := client.CurrentUserDevices(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected devices, got nil")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 devices, got %d", len(result))
	}
}

// TestCurrentUserTransferPlaybackEndpoint tests the CurrentUserTransferPlayback endpoint
func TestCurrentUserTransferPlaybackEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		deviceIDs, ok := body["device_ids"].([]interface{})
		if !ok || len(deviceIDs) == 0 {
			t.Error("expected device_ids array in request body")
		}

		w.WriteHeader(http.StatusOK)
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
	deviceIDs := []string{"device1"}
	err = client.CurrentUserTransferPlayback(ctx, deviceIDs, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserStartPlaybackEndpoint tests the CurrentUserStartPlayback endpoint
func TestCurrentUserStartPlaybackEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/me/player/play" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserStartPlayback(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserPausePlaybackEndpoint tests the CurrentUserPausePlayback endpoint
func TestCurrentUserPausePlaybackEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/me/player/pause" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserPausePlayback(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSeekToPositionEndpoint tests the CurrentUserSeekToPosition endpoint
func TestCurrentUserSeekToPositionEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/me/player/seek" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		position := r.URL.Query().Get("position_ms")
		if position != "30000" {
			t.Errorf("expected position_ms=30000, got %q", position)
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.SeekToPositionOptions{
		PositionMs: 30000,
	}
	err = client.CurrentUserSeekToPosition(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSetRepeatModeEndpoint tests the CurrentUserSetRepeatMode endpoint
func TestCurrentUserSetRepeatModeEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		state := r.URL.Query().Get("state")
		if state != "track" {
			t.Errorf("expected state=track, got %q", state)
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.SetRepeatModeOptions{
		State: "track",
	}
	err = client.CurrentUserSetRepeatMode(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSetVolumeEndpoint tests the CurrentUserSetVolume endpoint
func TestCurrentUserSetVolumeEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		volume := r.URL.Query().Get("volume_percent")
		if volume != "50" {
			t.Errorf("expected volume_percent=50, got %q", volume)
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.SetVolumeOptions{
		VolumePercent: 50,
	}
	err = client.CurrentUserSetVolume(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserToggleShuffleEndpoint tests the CurrentUserToggleShuffle endpoint
func TestCurrentUserToggleShuffleEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		state := r.URL.Query().Get("state")
		if state != "true" {
			t.Errorf("expected state=true, got %q", state)
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.ToggleShuffleOptions{
		State: true,
	}
	err = client.CurrentUserToggleShuffle(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSkipToNextEndpoint tests the CurrentUserSkipToNext endpoint
func TestCurrentUserSkipToNextEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/me/player/next" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserSkipToNext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSkipToPreviousEndpoint tests the CurrentUserSkipToPrevious endpoint
func TestCurrentUserSkipToPreviousEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/me/player/previous" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserSkipToPrevious(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserTopTracksEndpoint tests the CurrentUserTopTracks endpoint
func TestCurrentUserTopTracksEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/top/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "6b2oQwSGFkzsMtQruIWm2p", "name": "Top Track 1"},
			},
			"total":  1,
			"limit":  20,
			"offset": 0,
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
	result, err := client.CurrentUserTopTracks(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected top tracks response, got nil")
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 track, got %d", len(result.Items))
	}
}

// TestCurrentUserTopArtistsEndpoint tests the CurrentUserTopArtists endpoint
func TestCurrentUserTopArtistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/top/artists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "3jOstUTkEu2JkjvRdBA5Gu", "name": "Top Artist 1"},
			},
			"total":  1,
			"limit":  20,
			"offset": 0,
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
	result, err := client.CurrentUserTopArtists(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected top artists response, got nil")
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 artist, got %d", len(result.Items))
	}
}

// TestUserUnfollowArtistsEndpoint tests the UserUnfollowArtists endpoint
func TestUserUnfollowArtistsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["type"] != "artist" {
			t.Errorf("expected type=artist, got %v", body["type"])
		}

		w.WriteHeader(http.StatusOK)
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
	artistIDs := []string{"3jOstUTkEu2JkjvRdBA5Gu"}
	err = client.UserUnfollowArtists(ctx, artistIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestUserUnfollowUsersEndpoint tests the UserUnfollowUsers endpoint
func TestUserUnfollowUsersEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["type"] != "user" {
			t.Errorf("expected type=user, got %v", body["type"])
		}

		w.WriteHeader(http.StatusOK)
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
	userIDs := []string{"testuser"}
	err = client.UserUnfollowUsers(ctx, userIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserQueueEndpoint tests the CurrentUserQueue endpoint
func TestCurrentUserQueueEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/player/queue" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"currently_playing": map[string]interface{}{
				"id":   "6b2oQwSGFkzsMtQruIWm2p",
				"name": "Current Track",
			},
			"queue": []map[string]interface{}{
				{"id": "0Svkvt5I79wficMFgaqEQJ", "name": "Queued Track 1"},
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
	result, err := client.CurrentUserQueue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected queue response, got nil")
	}
}

// TestCurrentUserAddToQueueEndpoint tests the CurrentUserAddToQueue endpoint
func TestCurrentUserAddToQueueEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		uri := r.URL.Query().Get("uri")
		if uri == "" {
			t.Error("expected uri query parameter")
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserAddToQueue(ctx, "spotify:track:6b2oQwSGFkzsMtQruIWm2p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAvailableMarketsEndpoint tests the AvailableMarkets endpoint
func TestAvailableMarketsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"markets": []string{"US", "GB", "CA"},
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
	result, err := client.AvailableMarkets(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected markets, got nil")
	}

	if len(result) != 3 {
		t.Errorf("expected 3 markets, got %d", len(result))
	}
}

// TestCurrentUserSavedTracksWithOptions tests CurrentUserSavedTracks with options to improve coverage
func TestCurrentUserSavedTracksWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		offset := r.URL.Query().Get("offset")
		if offset != "5" {
			t.Errorf("expected offset=5, got %q", offset)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  10,
			"offset": 5,
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
	opts := &spotigo.SavedTracksOptions{
		Market: "US",
		Limit:  10,
		Offset: 5,
	}
	result, err := client.CurrentUserSavedTracks(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved tracks response, got nil")
	}
}

// TestCurrentUserSavedEpisodesWithOptions tests CurrentUserSavedEpisodes with options to improve coverage
func TestCurrentUserSavedEpisodesWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  10,
			"offset": 0,
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
	opts := &spotigo.SavedEpisodesOptions{
		Limit: 10,
	}
	result, err := client.CurrentUserSavedEpisodes(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved episodes response, got nil")
	}
}

// TestBrowseCategoriesWithOptions tests BrowseCategories with options to improve coverage
func TestBrowseCategoriesWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		country := r.URL.Query().Get("country")
		if country != "US" {
			t.Errorf("expected country=US, got %q", country)
		}

		locale := r.URL.Query().Get("locale")
		if locale != "en_US" {
			t.Errorf("expected locale=en_US, got %q", locale)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		offset := r.URL.Query().Get("offset")
		if offset != "5" {
			t.Errorf("expected offset=5, got %q", offset)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"categories": map[string]interface{}{
				"items":  []map[string]interface{}{},
				"total":  0,
				"limit":  10,
				"offset": 5,
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
	opts := &spotigo.BrowseCategoriesOptions{
		Country: "US",
		Locale:  "en_US",
		Limit:   10,
		Offset:  5,
	}
	result, err := client.BrowseCategories(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected categories response, got nil")
	}
}

// TestBrowseCategoryWithOptions tests BrowseCategory with options to improve coverage
func TestBrowseCategoryWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		country := r.URL.Query().Get("country")
		if country != "US" {
			t.Errorf("expected country=US, got %q", country)
		}

		locale := r.URL.Query().Get("locale")
		if locale != "en_US" {
			t.Errorf("expected locale=en_US, got %q", locale)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "pop",
			"name": "Pop",
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
	opts := &spotigo.BrowseCategoriesOptions{
		Country: "US",
		Locale:  "en_US",
	}
	category, err := client.BrowseCategory(ctx, "pop", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if category == nil {
		t.Fatal("expected category, got nil")
	}
}

// TestRecommendationGenreSeedsEndpoint tests the RecommendationGenreSeeds endpoint
func TestRecommendationGenreSeedsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/recommendations/available-genre-seeds" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"genres": []string{"acoustic", "afrobeat", "alt-rock", "alternative"},
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
	result, err := client.RecommendationGenreSeeds(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected genre seeds, got nil")
	}

	if len(result) == 0 {
		t.Error("expected at least one genre seed")
	}
}

// ============================================================================
// Audiobook Endpoints
// ============================================================================

// TestGetAudiobookEndpoint tests the GetAudiobook endpoint
func TestGetAudiobookEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audiobooks/7iHfbu1YPACw6oZPAFJtqe" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "7iHfbu1YPACw6oZPAFJtqe",
			"name": "Test Audiobook",
			"type": "audiobook",
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
	audiobook, err := client.GetAudiobook(ctx, "7iHfbu1YPACw6oZPAFJtqe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if audiobook == nil {
		t.Fatal("expected audiobook, got nil")
	}

	if audiobook.ID != "7iHfbu1YPACw6oZPAFJtqe" {
		t.Errorf("expected ID '7iHfbu1YPACw6oZPAFJtqe', got %q", audiobook.ID)
	}
}

// TestGetAudiobooksEndpoint tests the GetAudiobooks endpoint
func TestGetAudiobooksEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audiobooks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"audiobooks": []map[string]interface{}{
				{"id": "7iHfbu1YPACw6oZPAFJtqe", "name": "Audiobook 1", "type": "audiobook"},
				{"id": "8jHgcv2ZQBd7pZQGKtqrf", "name": "Audiobook 2", "type": "audiobook"},
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
	result, err := client.GetAudiobooks(ctx, []string{"7iHfbu1YPACw6oZPAFJtqe", "8jHgcv2ZQBd7pZQGKtqrf"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected audiobooks response, got nil")
	}

	if len(result.Audiobooks) != 2 {
		t.Errorf("expected 2 audiobooks, got %d", len(result.Audiobooks))
	}
}

// TestGetAudiobookChaptersEndpoint tests the GetAudiobookChapters endpoint
func TestGetAudiobookChaptersEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audiobooks/7iHfbu1YPACw6oZPAFJtqe/chapters" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "chapter1", "name": "Chapter 1", "type": "chapter"},
			},
			"total":  1,
			"limit":  20,
			"offset": 0,
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
	result, err := client.GetAudiobookChapters(ctx, "7iHfbu1YPACw6oZPAFJtqe", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected chapters response, got nil")
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 chapter, got %d", len(result.Items))
	}
}

// TestCurrentUserAddToQueueWithTrackID tests CurrentUserAddToQueue with track ID to improve coverage
func TestCurrentUserAddToQueueWithTrackID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		uri := r.URL.Query().Get("uri")
		if !strings.HasPrefix(uri, "spotify:track:") {
			t.Errorf("expected URI to start with spotify:track:, got %q", uri)
		}

		w.WriteHeader(http.StatusOK)
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
	// Test with track ID (should be converted to URI)
	err = client.CurrentUserAddToQueue(ctx, "6b2oQwSGFkzsMtQruIWm2p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserAddToQueueWithEpisodeURL tests CurrentUserAddToQueue with episode URL to improve coverage
func TestCurrentUserAddToQueueWithEpisodeURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		uri := r.URL.Query().Get("uri")
		if !strings.HasPrefix(uri, "spotify:episode:") {
			t.Errorf("expected URI to start with spotify:episode:, got %q", uri)
		}

		w.WriteHeader(http.StatusOK)
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
	// Test with episode URL (should be converted to episode URI)
	err = client.CurrentUserAddToQueue(ctx, "https://open.spotify.com/episode/4rOoJ6Egrf8K2IrywzwOMk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserRecentlyPlayedEndpoint tests the CurrentUserRecentlyPlayed endpoint
func TestCurrentUserRecentlyPlayedEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/player/recently-played" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"track": map[string]interface{}{
						"id":   "6b2oQwSGFkzsMtQruIWm2p",
						"name": "Recently Played Track",
					},
					"played_at": "2023-01-01T00:00:00Z",
				},
			},
			"cursors": map[string]interface{}{
				"after": "cursor_after",
			},
			"next":  "https://api.spotify.com/v1/me/player/recently-played?cursor=next",
			"href":  "https://api.spotify.com/v1/me/player/recently-played",
			"limit": 20,
			"total": 1,
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
	result, err := client.CurrentUserRecentlyPlayed(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected recently played response, got nil")
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
}

// TestCurrentUserRecentlyPlayedWithOptions tests CurrentUserRecentlyPlayed with options
func TestCurrentUserRecentlyPlayedWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		after := r.URL.Query().Get("after")
		if after == "" {
			t.Error("expected after parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"limit":  10,
			"total":  0,
			"cursors": map[string]interface{}{},
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
	after := int64(1234567890)
	opts := &spotigo.RecentlyPlayedOptions{
		Limit: 10,
		After: &after,
	}
	result, err := client.CurrentUserRecentlyPlayed(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected recently played response, got nil")
	}
}

// ============================================================================
// Pagination Helper Tests
// ============================================================================

// TestNextPaginationHelper tests the Next pagination helper
func TestNextPaginationHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The Next method uses the URL as-is, so we need to handle the full URL
		if !strings.Contains(r.URL.Path, "artists") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "artist1", "name": "Artist 1"},
			},
			"next":  nil,
			"total": 1,
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

	// Set API prefix to match server
	client.APIPrefix = server.URL + "/"

	ctx := context.Background()
	
	// Create a Paging object with next URL (use full URL from mock server)
	next := server.URL + "/artists?offset=20"
	paging := &spotigo.Paging[spotigo.Artist]{
		Next: &next,
	}

	result, err := client.Next(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestNextPaginationHelperNoNext tests Next when there's no next page
func TestNextPaginationHelperNoNext(t *testing.T) {
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
	
	// Create a Paging object with no next URL
	paging := &spotigo.Paging[spotigo.Artist]{
		Next: nil,
	}

	result, err := client.Next(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when no next page, got non-nil")
	}
}

// TestNextPaginationHelperWithMap tests Next with map[string]interface{}
func TestNextPaginationHelperWithMap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{},
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
	
	// Create a map with next URL (use full URL from mock server)
	nextURL := server.URL + "/artists?offset=20"
	paging := map[string]interface{}{
		"next": nextURL,
	}

	result, err := client.Next(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestPreviousPaginationHelper tests the Previous pagination helper
func TestPreviousPaginationHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{},
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
	
	// Use full URL from mock server
	prev := server.URL + "/artists?offset=0"
	paging := &spotigo.Paging[spotigo.Artist]{
		Previous: &prev,
	}

	result, err := client.Previous(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestNextGenericHelper tests the NextGeneric helper function
func TestNextGenericHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "artist1", "name": "Artist 1"},
			},
			"next":  nil,
			"total": 1,
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
	
	// Use full URL from mock server
	next := server.URL + "/artists?offset=20"
	paging := &spotigo.Paging[spotigo.Artist]{
		Next: &next,
	}

	result, err := spotigo.NextGeneric[spotigo.Artist](client, ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestNextCursorHelper tests the NextCursor helper function
func TestNextCursorHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "artist1", "name": "Artist 1"},
			},
			"cursors": map[string]interface{}{
				"after": "cursor_after",
			},
			"next":  nil,
			"total": 1,
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
	
	// Use full URL from mock server
	next := server.URL + "/me/following?cursor=next"
	paging := &spotigo.CursorPaging[spotigo.Artist]{
		Next: &next,
	}

	result, err := spotigo.NextCursor[spotigo.Artist](client, ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestPreviousGenericHelper tests the PreviousGeneric helper function
func TestPreviousGenericHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{},
			"total": 0,
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
	
	prev := server.URL + "/artists?offset=0"
	paging := &spotigo.Paging[spotigo.Artist]{
		Previous: &prev,
	}

	result, err := spotigo.PreviousGeneric[spotigo.Artist](client, ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestPreviousCursorHelper tests the PreviousCursor helper function
func TestPreviousCursorHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{},
			"cursors": map[string]interface{}{
				"before": "cursor_before",
			},
			"total": 0,
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
	
	prev := server.URL + "/me/following?cursor=prev"
	paging := &spotigo.CursorPaging[spotigo.Artist]{
		Previous: &prev,
	}

	result, err := spotigo.PreviousCursor[spotigo.Artist](client, ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestNextPaginationHelperUnsupportedType tests Next with unsupported type
func TestNextPaginationHelperUnsupportedType(t *testing.T) {
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
	
	// Use an unsupported type (int)
	paging := 123

	_, err = client.Next(ctx, paging)
	if err == nil {
		t.Fatal("expected error for unsupported pagination type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected error about unsupported type, got: %v", err)
	}
}

// TestPreviousPaginationHelperUnsupportedType tests Previous with unsupported type
func TestPreviousPaginationHelperUnsupportedType(t *testing.T) {
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
	
	// Use an unsupported type (int)
	paging := 123

	_, err = client.Previous(ctx, paging)
	if err == nil {
		t.Fatal("expected error for unsupported pagination type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected error about unsupported type, got: %v", err)
	}
}

// TestNextPaginationHelperWithEmptyNextURL tests Next with empty next URL in map
func TestNextPaginationHelperWithEmptyNextURL(t *testing.T) {
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
	
	// Create a map with empty next URL
	paging := map[string]interface{}{
		"next": "",
	}

	result, err := client.Next(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when next URL is empty, got non-nil")
	}
}

// TestNextPaginationHelperWithNilNextURL tests Next with nil next URL in map
func TestNextPaginationHelperWithNilNextURL(t *testing.T) {
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
	
	// Create a map with no next key
	paging := map[string]interface{}{
		"items": []interface{}{},
	}

	result, err := client.Next(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when no next key, got non-nil")
	}
}

// ============================================================================
// Edge Cases and Validation Tests
// ============================================================================

// TestCurrentUserSetVolumeValidation tests volume validation edge cases
func TestCurrentUserSetVolumeValidation(t *testing.T) {
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

	// Test: nil options
	err = client.CurrentUserSetVolume(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}

	// Test: volume < 0
	opts := &spotigo.SetVolumeOptions{
		VolumePercent: -1,
	}
	err = client.CurrentUserSetVolume(ctx, opts)
	if err == nil {
		t.Fatal("expected error for volume < 0, got nil")
	}

	// Test: volume > 100
	opts2 := &spotigo.SetVolumeOptions{
		VolumePercent: 101,
	}
	err = client.CurrentUserSetVolume(ctx, opts2)
	if err == nil {
		t.Fatal("expected error for volume > 100, got nil")
	}
}

// TestCurrentUserSeekToPositionValidation tests seek validation
func TestCurrentUserSeekToPositionValidation(t *testing.T) {
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

	// Test: nil options
	err = client.CurrentUserSeekToPosition(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}
}

// TestCurrentUserAddToQueueWithTrackURL tests CurrentUserAddToQueue with track URL
func TestCurrentUserAddToQueueWithTrackURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		uri := r.URL.Query().Get("uri")
		if !strings.HasPrefix(uri, "spotify:track:") {
			t.Errorf("expected URI to start with spotify:track:, got %q", uri)
		}

		w.WriteHeader(http.StatusOK)
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
	// Test with track URL (should be converted to track URI)
	err = client.CurrentUserAddToQueue(ctx, "https://open.spotify.com/track/6b2oQwSGFkzsMtQruIWm2p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserAddToQueueWithInvalidURI tests CurrentUserAddToQueue with invalid URI
func TestCurrentUserAddToQueueWithInvalidURI(t *testing.T) {
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
	// Test with invalid URI/URL/ID
	err = client.CurrentUserAddToQueue(ctx, "invalid_uri_12345")
	if err == nil {
		t.Fatal("expected error for invalid URI, got nil")
	}
}

// TestCurrentUserTransferPlaybackValidation tests transfer playback validation
func TestCurrentUserTransferPlaybackValidation(t *testing.T) {
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

	// Test: empty device IDs
	err = client.CurrentUserTransferPlayback(ctx, []string{}, nil)
	if err == nil {
		t.Fatal("expected error for empty device IDs, got nil")
	}
}

// TestCurrentUserSetRepeatModeValidation tests repeat mode validation
func TestCurrentUserSetRepeatModeValidation(t *testing.T) {
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

	// Test: nil options
	err = client.CurrentUserSetRepeatMode(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}
}

// TestCurrentUserToggleShuffleValidation tests toggle shuffle validation
func TestCurrentUserToggleShuffleValidation(t *testing.T) {
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

	// Test: nil options
	err = client.CurrentUserToggleShuffle(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil options, got nil")
	}
}

// TestPlaylistIsFollowingMaxLimit tests PlaylistIsFollowing max limit validation
func TestPlaylistIsFollowingMaxLimit(t *testing.T) {
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

	// Create 6 user IDs (exceeds max of 5)
	userIDs := make([]string, 6)
	for i := 0; i < 6; i++ {
		userIDs[i] = "user" + fmt.Sprintf("%d", i)
	}

	_, err = client.PlaylistIsFollowing(ctx, "2oCEWyyAPbZp9xhVSxZavx", userIDs)
	if err == nil {
		t.Fatal("expected error for exceeding max limit, got nil")
	}

	if !strings.Contains(err.Error(), "maximum 5") {
		t.Errorf("expected error about maximum 5, got: %v", err)
	}
}

// TestGetAudiobooksMaxLimit tests GetAudiobooks max limit validation
func TestGetAudiobooksMaxLimit(t *testing.T) {
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

	// Create 51 audiobook IDs (exceeds max of 50)
	audiobookIDs := make([]string, 51)
	for i := 0; i < 51; i++ {
		audiobookIDs[i] = "7iHfbu1YPACw6oZPAFJtqe"
	}

	_, err = client.GetAudiobooks(ctx, audiobookIDs)
	if err == nil {
		t.Fatal("expected error for exceeding max limit, got nil")
	}

	if !strings.Contains(err.Error(), "maximum 50") {
		t.Errorf("expected error about maximum 50, got: %v", err)
	}
}





// TestShowWithMarket tests Show with market parameter
func TestShowWithMarket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "4rOoJ6Egrf8K2IrywzwOMk",
			"name": "Test Show",
			"type": "show",
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
	show, err := client.Show(ctx, "4rOoJ6Egrf8K2IrywzwOMk", "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if show == nil {
		t.Fatal("expected show, got nil")
	}
}

// TestShowsWithMarket tests Shows with market parameter
func TestShowsWithMarket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"shows": []map[string]interface{}{
				{"id": "4rOoJ6Egrf8K2IrywzwOMk", "name": "Show 1"},
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
	result, err := client.Shows(ctx, []string{"4rOoJ6Egrf8K2IrywzwOMk"}, "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected shows response, got nil")
	}
}

// TestEpisodeWithMarket tests Episode with market parameter
func TestEpisodeWithMarket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "5AvwZVawapvyqJNqo3S7yF",
			"name": "Test Episode",
			"type": "episode",
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
	episode, err := client.Episode(ctx, "5AvwZVawapvyqJNqo3S7yF", "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if episode == nil {
		t.Fatal("expected episode, got nil")
	}
}

// TestEpisodesWithMarket tests Episodes with market parameter
func TestEpisodesWithMarket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"episodes": []map[string]interface{}{
				{"id": "5AvwZVawapvyqJNqo3S7yF", "name": "Episode 1"},
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
	result, err := client.Episodes(ctx, []string{"5AvwZVawapvyqJNqo3S7yF"}, "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected episodes response, got nil")
	}
}


// TestCurrentUserStartPlaybackWithURIs tests CurrentUserStartPlayback with URIs
func TestCurrentUserStartPlaybackWithURIs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		uris, ok := body["uris"].([]interface{})
		if !ok {
			t.Error("expected uris array in request body")
		}

		if len(uris) == 0 {
			t.Error("expected at least one URI")
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.StartPlaybackOptions{
		URIs: []string{"spotify:track:6b2oQwSGFkzsMtQruIWm2p"},
	}
	err = client.CurrentUserStartPlayback(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}


// TestCurrentUserSeekToPositionWithDeviceID tests CurrentUserSeekToPosition with device ID
func TestCurrentUserSeekToPositionWithDeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID != "device1" {
			t.Errorf("expected device_id=device1, got %q", deviceID)
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.SeekToPositionOptions{
		PositionMs: 30000,
		DeviceID:   "device1",
	}
	err = client.CurrentUserSeekToPosition(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSetRepeatModeWithDeviceID tests CurrentUserSetRepeatMode with device ID
func TestCurrentUserSetRepeatModeWithDeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID != "device1" {
			t.Errorf("expected device_id=device1, got %q", deviceID)
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.SetRepeatModeOptions{
		State:    "context",
		DeviceID: "device1",
	}
	err = client.CurrentUserSetRepeatMode(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSetVolumeWithDeviceID tests CurrentUserSetVolume with device ID
func TestCurrentUserSetVolumeWithDeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID != "device1" {
			t.Errorf("expected device_id=device1, got %q", deviceID)
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.SetVolumeOptions{
		VolumePercent: 75,
		DeviceID:      "device1",
	}
	err = client.CurrentUserSetVolume(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserToggleShuffleWithDeviceID tests CurrentUserToggleShuffle with device ID
func TestCurrentUserToggleShuffleWithDeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID != "device1" {
			t.Errorf("expected device_id=device1, got %q", deviceID)
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.ToggleShuffleOptions{
		State:    false,
		DeviceID: "device1",
	}
	err = client.CurrentUserToggleShuffle(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSkipToNextWithDeviceID tests CurrentUserSkipToNext with device ID
func TestCurrentUserSkipToNextWithDeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID != "device1" {
			t.Errorf("expected device_id=device1, got %q", deviceID)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserSkipToNext(ctx, "device1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserSkipToPreviousWithDeviceID tests CurrentUserSkipToPrevious with device ID
func TestCurrentUserSkipToPreviousWithDeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID != "device1" {
			t.Errorf("expected device_id=device1, got %q", deviceID)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserSkipToPrevious(ctx, "device1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPlaylistAddItemsWithPosition tests PlaylistAddItems with position
func TestPlaylistAddItemsWithPosition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		position, ok := body["position"].(float64)
		if !ok {
			t.Error("expected position in request body")
		}

		if position != 0 {
			t.Errorf("expected position=0, got %v", position)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"snapshot_id": "snapshot_id",
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
	items := []string{"spotify:track:6b2oQwSGFkzsMtQruIWm2p"}
	result, err := client.PlaylistAddItems(ctx, "2oCEWyyAPbZp9xhVSxZavx", items, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected snapshot ID, got nil")
	}
}

// TestPlaylistAddItemsMaxLimit tests PlaylistAddItems max limit validation
func TestPlaylistAddItemsMaxLimit(t *testing.T) {
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

	// Create 101 items (exceeds max of 100)
	items := make([]string, 101)
	for i := 0; i < 101; i++ {
		items[i] = "spotify:track:6b2oQwSGFkzsMtQruIWm2p"
	}

	_, err = client.PlaylistAddItems(ctx, "2oCEWyyAPbZp9xhVSxZavx", items)
	if err == nil {
		t.Fatal("expected error for exceeding max limit, got nil")
	}

	if !strings.Contains(err.Error(), "maximum 100") {
		t.Errorf("expected error about maximum 100, got: %v", err)
	}
}

// TestPlaylistAddItemsInvalidPosition tests PlaylistAddItems with invalid position
func TestPlaylistAddItemsInvalidPosition(t *testing.T) {
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

	items := []string{"spotify:track:6b2oQwSGFkzsMtQruIWm2p"}
	_, err = client.PlaylistAddItems(ctx, "2oCEWyyAPbZp9xhVSxZavx", items, -1)
	if err == nil {
		t.Fatal("expected error for negative position, got nil")
	}
}

// TestCurrentUserAddToQueueWithDeviceID tests CurrentUserAddToQueue with device ID
func TestCurrentUserAddToQueueWithDeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID != "device1" {
			t.Errorf("expected device_id=device1, got %q", deviceID)
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserAddToQueue(ctx, "spotify:track:6b2oQwSGFkzsMtQruIWm2p", "device1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserFollowPlaylistWithPublic tests CurrentUserFollowPlaylist with public parameter
func TestCurrentUserFollowPlaylistWithPublic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		public, ok := body["public"].(bool)
		if !ok {
			t.Error("expected public boolean in request body")
		}

		if !public {
			t.Error("expected public=true, got false")
		}

		w.WriteHeader(http.StatusOK)
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
	err = client.CurrentUserFollowPlaylist(ctx, "2oCEWyyAPbZp9xhVSxZavx", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCurrentUserTransferPlaybackWithPlay tests CurrentUserTransferPlayback with play option
func TestCurrentUserTransferPlaybackWithPlay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		play, ok := body["play"].(bool)
		if !ok {
			t.Error("expected play boolean in request body")
		}

		if !play {
			t.Error("expected play=true, got false")
		}

		w.WriteHeader(http.StatusOK)
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
	opts := &spotigo.TransferPlaybackOptions{
		Play: true,
	}
	err = client.CurrentUserTransferPlayback(ctx, []string{"device1"}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPreviousPaginationHelperWithMap tests Previous with map[string]interface{} to improve coverage
func TestPreviousPaginationHelperWithMap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{},
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
	
	// Use map[string]interface{} with previous URL
	prev := server.URL + "/artists?offset=0"
	paging := map[string]interface{}{
		"previous": prev,
	}

	result, err := client.Previous(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestPreviousPaginationHelperWithEmptyPrevURL tests Previous with empty previous URL in map
func TestPreviousPaginationHelperWithEmptyPrevURL(t *testing.T) {
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
	
	// Create a map with empty previous URL
	paging := map[string]interface{}{
		"previous": "",
	}

	result, err := client.Previous(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when previous URL is empty, got non-nil")
	}
}

// TestPreviousPaginationHelperWithNilPrevURL tests Previous with nil previous URL in map
func TestPreviousPaginationHelperWithNilPrevURL(t *testing.T) {
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
	
	// Create a map with no previous key
	paging := map[string]interface{}{
		"items": []interface{}{},
	}

	result, err := client.Previous(ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when no previous key, got non-nil")
	}
}

// TestNextGenericHelperNoNext tests NextGeneric when there's no next page
func TestNextGenericHelperNoNext(t *testing.T) {
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
	
	// Create paging with nil next
	paging := &spotigo.Paging[spotigo.Artist]{
		Next: nil,
	}

	result, err := spotigo.NextGeneric[spotigo.Artist](client, ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when next is nil, got non-nil")
	}
}

// TestPreviousGenericHelperNoPrevious tests PreviousGeneric when there's no previous page
func TestPreviousGenericHelperNoPrevious(t *testing.T) {
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
	
	// Create paging with nil previous
	paging := &spotigo.Paging[spotigo.Artist]{
		Previous: nil,
	}

	result, err := spotigo.PreviousGeneric[spotigo.Artist](client, ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when previous is nil, got non-nil")
	}
}

// TestNextCursorHelperNoNext tests NextCursor when there's no next page
func TestNextCursorHelperNoNext(t *testing.T) {
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
	
	// Create cursor paging with nil next
	paging := &spotigo.CursorPaging[spotigo.Artist]{
		Next: nil,
	}

	result, err := spotigo.NextCursor[spotigo.Artist](client, ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when next is nil, got non-nil")
	}
}

// TestPreviousCursorHelperNoPrevious tests PreviousCursor when there's no previous page
func TestPreviousCursorHelperNoPrevious(t *testing.T) {
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
	
	// Create cursor paging with nil previous
	paging := &spotigo.CursorPaging[spotigo.Artist]{
		Previous: nil,
	}

	result, err := spotigo.PreviousCursor[spotigo.Artist](client, ctx, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result when previous is nil, got non-nil")
	}
}













// TestCurrentUserSavedEpisodesNilOptions tests CurrentUserSavedEpisodes with nil options to improve coverage
func TestCurrentUserSavedEpisodesNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "20" {
			t.Errorf("expected default limit=20, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  20,
			"offset": 0,
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
	// Test with nil options (should use defaults)
	result, err := client.CurrentUserSavedEpisodes(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved episodes response, got nil")
	}
}

// TestCurrentUserSavedEpisodesLimitCapping tests CurrentUserSavedEpisodes with limit > 50 to improve coverage
func TestCurrentUserSavedEpisodesLimitCapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "50" {
			t.Errorf("expected limit to be capped at 50, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  50,
			"offset": 0,
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
	opts := &spotigo.SavedEpisodesOptions{
		Limit: 100, // Should be capped at 50
	}
	result, err := client.CurrentUserSavedEpisodes(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved episodes response, got nil")
	}
}

// TestCurrentUserSavedEpisodesWithOffset tests CurrentUserSavedEpisodes with offset to improve coverage
func TestCurrentUserSavedEpisodesWithOffset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		if offset != "10" {
			t.Errorf("expected offset=10, got %q", offset)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  20,
			"offset": 10,
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
	opts := &spotigo.SavedEpisodesOptions{
		Offset: 10,
	}
	result, err := client.CurrentUserSavedEpisodes(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected saved episodes response, got nil")
	}
}

// TestBrowseFeaturedPlaylistsWithAllOptions tests BrowseFeaturedPlaylists with all options to improve coverage
func TestBrowseFeaturedPlaylistsWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		country := r.URL.Query().Get("country")
		if country != "US" {
			t.Errorf("expected country=US, got %q", country)
		}

		locale := r.URL.Query().Get("locale")
		if locale != "en_US" {
			t.Errorf("expected locale=en_US, got %q", locale)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		offset := r.URL.Query().Get("offset")
		if offset != "5" {
			t.Errorf("expected offset=5, got %q", offset)
		}

		timestamp := r.URL.Query().Get("timestamp")
		if timestamp != "2024-01-01T00:00:00" {
			t.Errorf("expected timestamp, got %q", timestamp)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Featured Playlists",
			"playlists": map[string]interface{}{
				"items":  []map[string]interface{}{},
				"total":  0,
				"limit":  10,
				"offset": 5,
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
	opts := &spotigo.FeaturedPlaylistsOptions{
		Country:   "US",
		Locale:    "en_US",
		Limit:     10,
		Offset:    5,
		Timestamp: "2024-01-01T00:00:00",
	}
	result, err := client.BrowseFeaturedPlaylists(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected featured playlists response, got nil")
	}
}

// TestBrowseFeaturedPlaylistsNilOptions tests BrowseFeaturedPlaylists with nil options to improve coverage
func TestBrowseFeaturedPlaylistsNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "20" {
			t.Errorf("expected default limit=20, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Featured Playlists",
			"playlists": map[string]interface{}{
				"items":  []map[string]interface{}{},
				"total":  0,
				"limit":  20,
				"offset": 0,
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
	// Test with nil options (should use defaults)
	result, err := client.BrowseFeaturedPlaylists(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected featured playlists response, got nil")
	}
}

// TestBrowseFeaturedPlaylistsLimitCapping tests BrowseFeaturedPlaylists with limit > 50 to improve coverage
func TestBrowseFeaturedPlaylistsLimitCapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "50" {
			t.Errorf("expected limit to be capped at 50, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Featured Playlists",
			"playlists": map[string]interface{}{
				"items":  []map[string]interface{}{},
				"total":  0,
				"limit":  50,
				"offset": 0,
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
	opts := &spotigo.FeaturedPlaylistsOptions{
		Limit: 100, // Should be capped at 50
	}
	result, err := client.BrowseFeaturedPlaylists(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected featured playlists response, got nil")
	}
}

// TestPlaylistTracksWithAllOptions tests PlaylistTracks with all options to improve coverage
func TestPlaylistTracksWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := r.URL.Query().Get("fields")
		if fields != "items(track(name))" {
			t.Errorf("expected fields, got %q", fields)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		offset := r.URL.Query().Get("offset")
		if offset != "5" {
			t.Errorf("expected offset=5, got %q", offset)
		}

		market := r.URL.Query().Get("market")
		if market != "US" {
			t.Errorf("expected market=US, got %q", market)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  10,
			"offset": 5,
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
	opts := &spotigo.PlaylistTracksOptions{
		Fields: "items(track(name))",
		Limit:  10,
		Offset: 5,
		Market: "US",
	}
	result, err := client.PlaylistTracks(ctx, "2oCEWyyAPbZp9xhVSxZavx", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected playlist tracks response, got nil")
	}
}

// TestPlaylistTracksNilOptions tests PlaylistTracks with nil options to improve coverage
func TestPlaylistTracksNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "100" {
			t.Errorf("expected default limit=100, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  100,
			"offset": 0,
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
	// Test with nil options (should use defaults)
	result, err := client.PlaylistTracks(ctx, "2oCEWyyAPbZp9xhVSxZavx", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected playlist tracks response, got nil")
	}
}

// TestPlaylistTracksLimitCapping tests PlaylistTracks with limit > 100 to improve coverage
func TestPlaylistTracksLimitCapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "100" {
			t.Errorf("expected limit to be capped at 100, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  100,
			"offset": 0,
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
	opts := &spotigo.PlaylistTracksOptions{
		Limit: 200, // Should be capped at 100
	}
	result, err := client.PlaylistTracks(ctx, "2oCEWyyAPbZp9xhVSxZavx", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected playlist tracks response, got nil")
	}
}

// TestCurrentUserTopTracksWithAllOptions tests CurrentUserTopTracks with all options to improve coverage
func TestCurrentUserTopTracksWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeRange := r.URL.Query().Get("time_range")
		if timeRange != "medium_term" {
			t.Errorf("expected time_range=medium_term, got %q", timeRange)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		offset := r.URL.Query().Get("offset")
		if offset != "5" {
			t.Errorf("expected offset=5, got %q", offset)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  10,
			"offset": 5,
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
	opts := &spotigo.TopItemsOptions{
		TimeRange: "medium_term",
		Limit:     10,
		Offset:    5,
	}
	result, err := client.CurrentUserTopTracks(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected top tracks response, got nil")
	}
}

// TestCurrentUserTopTracksNilOptions tests CurrentUserTopTracks with nil options to improve coverage
func TestCurrentUserTopTracksNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "20" {
			t.Errorf("expected default limit=20, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  20,
			"offset": 0,
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
	// Test with nil options (should use defaults)
	result, err := client.CurrentUserTopTracks(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected top tracks response, got nil")
	}
}

// TestCurrentUserTopTracksLimitCapping tests CurrentUserTopTracks with limit > 50 to improve coverage
func TestCurrentUserTopTracksLimitCapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "50" {
			t.Errorf("expected limit to be capped at 50, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  50,
			"offset": 0,
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
	opts := &spotigo.TopItemsOptions{
		Limit: 100, // Should be capped at 50
	}
	result, err := client.CurrentUserTopTracks(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected top tracks response, got nil")
	}
}

// TestCurrentUserTopArtistsWithAllOptions tests CurrentUserTopArtists with all options to improve coverage
func TestCurrentUserTopArtistsWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeRange := r.URL.Query().Get("time_range")
		if timeRange != "long_term" {
			t.Errorf("expected time_range=long_term, got %q", timeRange)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":  []map[string]interface{}{},
			"total":  0,
			"limit":  10,
			"offset": 0,
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
	opts := &spotigo.TopItemsOptions{
		TimeRange: "long_term",
		Limit:     10,
	}
	result, err := client.CurrentUserTopArtists(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected top artists response, got nil")
	}
}

// TestBrowseNewReleasesWithAllOptions tests BrowseNewReleases with all options to improve coverage
func TestBrowseNewReleasesWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		country := r.URL.Query().Get("country")
		if country != "US" {
			t.Errorf("expected country=US, got %q", country)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %q", limit)
		}

		offset := r.URL.Query().Get("offset")
		if offset != "5" {
			t.Errorf("expected offset=5, got %q", offset)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"albums": map[string]interface{}{
				"items":  []map[string]interface{}{},
				"total":  0,
				"limit":  10,
				"offset": 5,
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
	opts := &spotigo.NewReleasesOptions{
		Country: "US",
		Limit:   10,
		Offset:  5,
	}
	result, err := client.BrowseNewReleases(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected new releases response, got nil")
	}
}

// TestBrowseNewReleasesNilOptions tests BrowseNewReleases with nil options to improve coverage
func TestBrowseNewReleasesNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "20" {
			t.Errorf("expected default limit=20, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"albums": map[string]interface{}{
				"items":  []map[string]interface{}{},
				"total":  0,
				"limit":  20,
				"offset": 0,
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
	// Test with nil options (should use defaults)
	result, err := client.BrowseNewReleases(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected new releases response, got nil")
	}
}

// TestBrowseNewReleasesLimitCapping tests BrowseNewReleases with limit > 50 to improve coverage
func TestBrowseNewReleasesLimitCapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "50" {
			t.Errorf("expected limit to be capped at 50, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"albums": map[string]interface{}{
				"items":  []map[string]interface{}{},
				"total":  0,
				"limit":  50,
				"offset": 0,
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
	opts := &spotigo.NewReleasesOptions{
		Limit: 100, // Should be capped at 50
	}
	result, err := client.BrowseNewReleases(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected new releases response, got nil")
	}
}
