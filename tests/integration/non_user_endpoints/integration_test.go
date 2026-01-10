package non_user_endpoints

import (
	"context"
	"testing"

	spotigotests "github.com/sv4u/spotigo/tests"
)

// Test data from Spotipy tests
const (
	CreepURI  = "spotify:track:6b2oQwSGFkzsMtQruIWm2p"
	CreepID   = "6b2oQwSGFkzsMtQruIWm2p"
	CreepURL  = "http://open.spotify.com/track/6b2oQwSGFkzsMtQruIWm2p"
	WeezerURI = "spotify:artist:3jOstUTkEu2JkjvRdBA5Gu"
	WeezerID  = "3jOstUTkEu2JkjvRdBA5Gu"
	PinkertonURI = "spotify:album:04xe676vyiTeYNXw15o9jT"
	PinkertonID  = "04xe676vyiTeYNXw15o9jT"
)

func TestTrackByURI(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	track, err := client.Track(ctx, CreepURI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if track == nil {
		t.Fatal("expected track, got nil")
	}

	if track.Name != "Creep" {
		t.Errorf("expected name 'Creep', got %q", track.Name)
	}
}

func TestTrackByID(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	track, err := client.Track(ctx, CreepID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if track == nil {
		t.Fatal("expected track, got nil")
	}

	if track.Name != "Creep" {
		t.Errorf("expected name 'Creep', got %q", track.Name)
	}
}

func TestTrackByURL(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	track, err := client.Track(ctx, CreepURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if track == nil {
		t.Fatal("expected track, got nil")
	}

	if track.Name != "Creep" {
		t.Errorf("expected name 'Creep', got %q", track.Name)
	}
}

func TestTracksMultiple(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Tracks(ctx, []string{CreepURL, CreepURI})
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

func TestArtistByURI(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	artist, err := client.Artist(ctx, WeezerURI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if artist == nil {
		t.Fatal("expected artist, got nil")
	}

	if artist.Name != "Weezer" {
		t.Errorf("expected name 'Weezer', got %q", artist.Name)
	}
}

func TestAlbumByURI(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	album, err := client.Album(ctx, PinkertonURI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if album == nil {
		t.Fatal("expected album, got nil")
	}

	if album.Name != "Pinkerton" {
		t.Errorf("expected name 'Pinkerton', got %q", album.Name)
	}
}

func TestAlbumTracks(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	result, err := client.AlbumTracks(ctx, PinkertonURI, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected tracks, got nil")
	}

	if len(result.Items) == 0 {
		t.Error("expected at least one track, got 0")
	}
}

func TestArtistAlbums(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	result, err := client.ArtistAlbums(ctx, WeezerURI, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected albums, got nil")
	}

	if len(result.Items) == 0 {
		t.Error("expected at least one album, got 0")
	}
}

func TestArtistTopTracks(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	result, err := client.ArtistTopTracks(ctx, WeezerURI, "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected tracks, got nil")
	}

	if len(result.Tracks) == 0 {
		t.Error("expected at least one track, got 0")
	}
}

func TestSearch(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Search(ctx, "weezer", "artist", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected search result, got nil")
	}

	if result.Artists == nil {
		t.Fatal("expected artists in search result")
	}

	if len(result.Artists.Items) == 0 {
		t.Error("expected at least one artist, got 0")
	}
}

func TestTrackInvalidID(t *testing.T) {
	spotigotests.SkipIfNoCredentials(t)

	client, err := spotigotests.NewTestClient(t)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	ctx := context.Background()
	_, err = client.Track(ctx, "BAD_ID")
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}
