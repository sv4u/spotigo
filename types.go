package spotigo

// Type definitions for Spotify Web API responses
// All types match the Spotify API JSON structure exactly

// Paging represents a paginated response (offset-based)
type Paging[T any] struct {
	Href     string  `json:"href"`
	Items    []T     `json:"items"`
	Limit    int     `json:"limit"`
	Next     *string `json:"next"`
	Offset   int     `json:"offset"`
	Previous *string `json:"previous"`
	Total    int     `json:"total"`
}

// GetNext returns the next page URL
func (p *Paging[T]) GetNext() *string {
	return p.Next
}

// GetPrevious returns the previous page URL
func (p *Paging[T]) GetPrevious() *string {
	return p.Previous
}

// CursorPaging represents a cursor-based paginated response
type CursorPaging[T any] struct {
	Href     string   `json:"href"`
	Items    []T      `json:"items"`
	Limit    int      `json:"limit"`
	Next     *string  `json:"next"`
	Previous *string  `json:"previous,omitempty"` // Some cursor pagination has previous
	Cursors  *Cursors `json:"cursors"`
	Total    int      `json:"total"`
}

// GetNext returns the next page URL
func (p *CursorPaging[T]) GetNext() *string {
	return p.Next
}

// GetPrevious returns the previous page URL
func (p *CursorPaging[T]) GetPrevious() *string {
	return p.Previous
}

// Cursors represents pagination cursors
type Cursors struct {
	After  *string `json:"after"`
	Before *string `json:"before"`
}

// Common types used across multiple entities

// Image represents an image
type Image struct {
	URL    string `json:"url"`
	Height *int   `json:"height,omitempty"`
	Width  *int   `json:"width,omitempty"`
}

// ExternalURLs represents external URLs
type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

// ExternalIDs represents external identifiers
type ExternalIDs struct {
	ISRC *string `json:"isrc,omitempty"`
	EAN  *string `json:"ean,omitempty"`
	UPC  *string `json:"upc,omitempty"`
}

// Restrictions represents content restrictions
type Restrictions struct {
	Reason string `json:"reason,omitempty"`
}

// Followers represents follower information
type Followers struct {
	Href  *string `json:"href,omitempty"`
	Total int     `json:"total"`
}

// SimplifiedArtist represents a simplified artist object
type SimplifiedArtist struct {
	ExternalURLs *ExternalURLs `json:"external_urls"`
	Href         string        `json:"href"`
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	URI          string        `json:"uri"`
}

// Artist represents a full artist object
type Artist struct {
	ExternalURLs *ExternalURLs `json:"external_urls"`
	Followers    *Followers    `json:"followers"`
	Genres       []string      `json:"genres"`
	Href         string        `json:"href"`
	ID           string        `json:"id"`
	Images       []Image       `json:"images"`
	Name         string        `json:"name"`
	Popularity   int           `json:"popularity"`
	Type         string        `json:"type"`
	URI          string        `json:"uri"`
}

// SimplifiedAlbum represents a simplified album object
type SimplifiedAlbum struct {
	AlbumType            string        `json:"album_type"`
	Artists              []Artist      `json:"artists"`
	AvailableMarkets     []string      `json:"available_markets"`
	ExternalURLs         *ExternalURLs `json:"external_urls"`
	Href                 string        `json:"href"`
	ID                   string        `json:"id"`
	Images               []Image       `json:"images"`
	Name                 string        `json:"name"`
	ReleaseDate          string        `json:"release_date"`
	ReleaseDatePrecision string        `json:"release_date_precision"`
	Restrictions         *Restrictions `json:"restrictions,omitempty"`
	TotalTracks          int           `json:"total_tracks"`
	Type                 string        `json:"type"`
	URI                  string        `json:"uri"`
}

// TrackLink represents a link to another track
type TrackLink struct {
	ExternalURLs *ExternalURLs `json:"external_urls"`
	Href         string        `json:"href"`
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	URI          string        `json:"uri"`
}

// Track represents a full track object
type Track struct {
	Album            *SimplifiedAlbum `json:"album"`
	Artists          []Artist         `json:"artists"`
	AvailableMarkets []string         `json:"available_markets"`
	DiscNumber       int              `json:"disc_number"`
	DurationMs       int              `json:"duration_ms"`
	Explicit         bool             `json:"explicit"`
	ExternalIDs      *ExternalIDs     `json:"external_ids"`
	ExternalURLs     *ExternalURLs    `json:"external_urls"`
	Href             string           `json:"href"`
	ID               string           `json:"id"`
	IsPlayable       *bool            `json:"is_playable,omitempty"`
	LinkedFrom       *TrackLink       `json:"linked_from,omitempty"`
	Restrictions     *Restrictions    `json:"restrictions,omitempty"`
	Name             string           `json:"name"`
	Popularity       int              `json:"popularity"`
	PreviewURL       *string          `json:"preview_url,omitempty"`
	TrackNumber      int              `json:"track_number"`
	Type             string           `json:"type"`
	URI              string           `json:"uri"`
	IsLocal          bool             `json:"is_local"`
}

// SimplifiedTrack represents a simplified track object
type SimplifiedTrack struct {
	Artists          []SimplifiedArtist `json:"artists"`
	AvailableMarkets []string           `json:"available_markets"`
	DiscNumber       int                `json:"disc_number"`
	DurationMs       int                `json:"duration_ms"`
	Explicit         bool               `json:"explicit"`
	ExternalURLs     *ExternalURLs      `json:"external_urls"`
	Href             string             `json:"href"`
	ID               string             `json:"id"`
	IsPlayable       *bool              `json:"is_playable,omitempty"`
	LinkedFrom       *TrackLink         `json:"linked_from,omitempty"`
	Restrictions     *Restrictions      `json:"restrictions,omitempty"`
	Name             string             `json:"name"`
	PreviewURL       *string            `json:"preview_url,omitempty"`
	TrackNumber      int                `json:"track_number"`
	Type             string             `json:"type"`
	URI              string             `json:"uri"`
	IsLocal          bool               `json:"is_local"`
}

// Response types for multiple items

// TracksResponse represents a response with multiple tracks
type TracksResponse struct {
	Tracks []Track `json:"tracks"`
}

// ArtistsResponse represents a response with multiple artists
type ArtistsResponse struct {
	Artists []Artist `json:"artists"`
}

// Album represents a full album object
type Album struct {
	AlbumType            string                   `json:"album_type"`
	Artists              []Artist                 `json:"artists"`
	AvailableMarkets     []string                 `json:"available_markets"`
	Copyrights           []Copyright              `json:"copyrights"`
	ExternalIDs          *ExternalIDs             `json:"external_ids"`
	ExternalURLs         *ExternalURLs            `json:"external_urls"`
	Genres               []string                 `json:"genres"`
	Href                 string                   `json:"href"`
	ID                   string                   `json:"id"`
	Images               []Image                  `json:"images"`
	Label                string                   `json:"label"`
	Name                 string                   `json:"name"`
	Popularity           int                      `json:"popularity"`
	ReleaseDate          string                   `json:"release_date"`
	ReleaseDatePrecision string                   `json:"release_date_precision"`
	Restrictions         *Restrictions            `json:"restrictions,omitempty"`
	Tracks               *Paging[SimplifiedTrack] `json:"tracks"`
	TotalTracks          int                      `json:"total_tracks"`
	Type                 string                   `json:"type"`
	URI                  string                   `json:"uri"`
}

// Copyright represents copyright information
type Copyright struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// AlbumsResponse represents a response with multiple albums
type AlbumsResponse struct {
	Albums []Album `json:"albums"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Tracks     *Paging[Track]               `json:"tracks,omitempty"`
	Artists    *Paging[Artist]              `json:"artists,omitempty"`
	Albums     *Paging[SimplifiedAlbum]     `json:"albums,omitempty"`
	Playlists  *Paging[SimplifiedPlaylist]  `json:"playlists,omitempty"`
	Shows      *Paging[SimplifiedShow]      `json:"shows,omitempty"`
	Episodes   *Paging[SimplifiedEpisode]   `json:"episodes,omitempty"`
	Audiobooks *Paging[SimplifiedAudiobook] `json:"audiobooks,omitempty"`
}

// SimplifiedPlaylist represents a simplified playlist object
type SimplifiedPlaylist struct {
	Collaborative bool               `json:"collaborative"`
	Description   *string            `json:"description"`
	ExternalURLs  *ExternalURLs      `json:"external_urls"`
	Href          string             `json:"href"`
	ID            string             `json:"id"`
	Images        []Image            `json:"images"`
	Name          string             `json:"name"`
	Owner         *PublicUser        `json:"owner"`
	Public        *bool              `json:"public"`
	SnapshotID    string             `json:"snapshot_id"`
	Tracks        *PlaylistTracksRef `json:"tracks"`
	Type          string             `json:"type"`
	URI           string             `json:"uri"`
}

// Playlist represents a full playlist object
type Playlist struct {
	SimplifiedPlaylist
	Followers   *Followers `json:"followers"`
	Description *string    `json:"description"`
}

// PlaylistTracksRef represents a reference to playlist tracks
type PlaylistTracksRef struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

// PlaylistTrack represents a track in a playlist
type PlaylistTrack struct {
	AddedAt string      `json:"added_at"`
	AddedBy *PublicUser `json:"added_by"`
	IsLocal bool        `json:"is_local"`
	Track   interface{} `json:"track"` // Can be Track or Episode
}

// PlaylistItem represents an item in a playlist (track or episode)
type PlaylistItem struct {
	AddedAt string      `json:"added_at"`
	AddedBy *PublicUser `json:"added_by"`
	IsLocal bool        `json:"is_local"`
	Track   *Track      `json:"track,omitempty"`
	Episode *Episode    `json:"episode,omitempty"`
}

// PublicUser represents a public user profile
type PublicUser struct {
	DisplayName  *string       `json:"display_name"`
	ExternalURLs *ExternalURLs `json:"external_urls"`
	Followers    *Followers    `json:"followers"`
	Href         string        `json:"href"`
	ID           string        `json:"id"`
	Images       []Image       `json:"images"`
	Type         string        `json:"type"`
	URI          string        `json:"uri"`
}

// PlaylistSnapshotID represents a playlist snapshot ID response
type PlaylistSnapshotID struct {
	SnapshotID string `json:"snapshot_id"`
}

// PlaylistItemToRemove represents an item to remove from a playlist
// PlaylistItemToRemove represents an item to remove from a playlist
// Format matches Spotify Web API specification: {"uri": "...", "positions": [...]}
// uri: Spotify URI of the track/episode to remove
// positions: Optional array of positions where the item appears (0-based indices)
type PlaylistItemToRemove struct {
	URI       string `json:"uri"`
	Positions []int  `json:"positions,omitempty"`
}

// SimplifiedShow represents a simplified show object
type SimplifiedShow struct {
	AvailableMarkets   []string      `json:"available_markets"`
	Copyrights         []Copyright   `json:"copyrights"`
	Description        string        `json:"description"`
	Explicit           bool          `json:"explicit"`
	ExternalURLs       *ExternalURLs `json:"external_urls"`
	Href               string        `json:"href"`
	ID                 string        `json:"id"`
	Images             []Image       `json:"images"`
	IsExternallyHosted bool          `json:"is_externally_hosted"`
	Languages          []string      `json:"languages"`
	MediaType          string        `json:"media_type"`
	Name               string        `json:"name"`
	Publisher          string        `json:"publisher"`
	Type               string        `json:"type"`
	URI                string        `json:"uri"`
	TotalEpisodes      int           `json:"total_episodes"`
}

// SimplifiedEpisode represents a simplified episode object
type SimplifiedEpisode struct {
	AudioPreviewURL      *string       `json:"audio_preview_url"`
	Description          string        `json:"description"`
	DurationMs           int           `json:"duration_ms"`
	Explicit             bool          `json:"explicit"`
	ExternalURLs         *ExternalURLs `json:"external_urls"`
	Href                 string        `json:"href"`
	ID                   string        `json:"id"`
	Images               []Image       `json:"images"`
	IsExternallyHosted   bool          `json:"is_externally_hosted"`
	IsPlayable           bool          `json:"is_playable"`
	Language             *string       `json:"language"`
	Languages            []string      `json:"languages"`
	Name                 string        `json:"name"`
	ReleaseDate          string        `json:"release_date"`
	ReleaseDatePrecision string        `json:"release_date_precision"`
	Restrictions         *Restrictions `json:"restrictions,omitempty"`
	Type                 string        `json:"type"`
	URI                  string        `json:"uri"`
}

// Episode represents a full episode object
type Episode struct {
	SimplifiedEpisode
	Show *SimplifiedShow `json:"show"`
}

// SimplifiedAudiobook represents a simplified audiobook object
type SimplifiedAudiobook struct {
	Authors          []Author      `json:"authors"`
	AvailableMarkets []string      `json:"available_markets"`
	Copyrights       []Copyright   `json:"copyrights"`
	Description      string        `json:"description"`
	Edition          *string       `json:"edition"`
	Explicit         bool          `json:"explicit"`
	ExternalURLs     *ExternalURLs `json:"external_urls"`
	Href             string        `json:"href"`
	ID               string        `json:"id"`
	Images           []Image       `json:"images"`
	Languages        []string      `json:"languages"`
	MediaType        string        `json:"media_type"`
	Name             string        `json:"name"`
	Narrators        []Narrator    `json:"narrators"`
	Publisher        string        `json:"publisher"`
	Type             string        `json:"type"`
	URI              string        `json:"uri"`
	TotalChapters    int           `json:"total_chapters"`
}

// Author represents an author
type Author struct {
	Name string `json:"name"`
}

// Narrator represents a narrator
type Narrator struct {
	Name string `json:"name"`
}

// User represents a full user profile
type User struct {
	Country         *string                  `json:"country"`
	DisplayName     *string                  `json:"display_name"`
	Email           *string                  `json:"email"`
	ExplicitContent *ExplicitContentSettings `json:"explicit_content"`
	ExternalURLs    *ExternalURLs            `json:"external_urls"`
	Followers       *Followers               `json:"followers"`
	Href            string                   `json:"href"`
	ID              string                   `json:"id"`
	Images          []Image                  `json:"images"`
	Product         *string                  `json:"product"`
	Type            string                   `json:"type"`
	URI             string                   `json:"uri"`
}

// ExplicitContentSettings represents explicit content settings
type ExplicitContentSettings struct {
	FilterEnabled bool `json:"filter_enabled"`
	FilterLocked  bool `json:"filter_locked"`
}

// Show represents a full show object
type Show struct {
	SimplifiedShow
	Episodes *Paging[SimplifiedEpisode] `json:"episodes"`
}

// ShowsResponse represents a response with multiple shows
type ShowsResponse struct {
	Shows []Show `json:"shows"`
}

// EpisodesResponse represents a response with multiple episodes
type EpisodesResponse struct {
	Episodes []Episode `json:"episodes"`
}

// Audiobook represents a full audiobook object
type Audiobook struct {
	SimplifiedAudiobook
	Chapters *Paging[Chapter] `json:"chapters"`
}

// AudiobooksResponse represents a response with multiple audiobooks
type AudiobooksResponse struct {
	Audiobooks []Audiobook `json:"audiobooks"`
}

// Chapter represents a chapter object
type Chapter struct {
	Audiobook            *SimplifiedAudiobook `json:"audiobook,omitempty"`
	AudioPreviewURL      *string              `json:"audio_preview_url"`
	AvailableMarkets     []string             `json:"available_markets"`
	ChapterNumber        int                  `json:"chapter_number"`
	Description          string               `json:"description"`
	HTMLDescription      string               `json:"html_description"`
	DurationMs           int                  `json:"duration_ms"`
	Explicit             bool                 `json:"explicit"`
	ExternalURLs         *ExternalURLs        `json:"external_urls"`
	Href                 string               `json:"href"`
	ID                   string               `json:"id"`
	Images               []Image              `json:"images"`
	IsPlayable           bool                 `json:"is_playable"`
	Languages            []string             `json:"languages"`
	Name                 string               `json:"name"`
	ReleaseDate          string               `json:"release_date"`
	ReleaseDatePrecision string               `json:"release_date_precision"`
	ResumePoint          *ResumePoint         `json:"resume_point,omitempty"`
	Type                 string               `json:"type"`
	URI                  string               `json:"uri"`
	Restrictions         *Restrictions        `json:"restrictions,omitempty"`
}

// ResumePoint represents a resume point
type ResumePoint struct {
	FullyPlayed      bool `json:"fully_played"`
	ResumePositionMs int  `json:"resume_position_ms"`
}

// SavedTrack represents a saved track
type SavedTrack struct {
	AddedAt string `json:"added_at"`
	Track   Track  `json:"track"`
}

// SavedAlbum represents a saved album
type SavedAlbum struct {
	AddedAt string `json:"added_at"`
	Album   Album  `json:"album"`
}

// SavedEpisode represents a saved episode
type SavedEpisode struct {
	AddedAt string  `json:"added_at"`
	Episode Episode `json:"episode"`
}

// SavedShow represents a saved show
type SavedShow struct {
	AddedAt string `json:"added_at"`
	Show    Show   `json:"show"`
}

// AudioFeatures represents audio features for a track
type AudioFeatures struct {
	Danceability     float64 `json:"danceability"`
	Energy           float64 `json:"energy"`
	Key              int     `json:"key"`
	Loudness         float64 `json:"loudness"`
	Mode             int     `json:"mode"`
	Speechiness      float64 `json:"speechiness"`
	Acousticness     float64 `json:"acousticness"`
	Instrumentalness float64 `json:"instrumentalness"`
	Liveness         float64 `json:"liveness"`
	Valence          float64 `json:"valence"`
	Tempo            float64 `json:"tempo"`
	Type             string  `json:"type"`
	ID               string  `json:"id"`
	URI              string  `json:"uri"`
	TrackHref        string  `json:"track_href"`
	AnalysisURL      string  `json:"analysis_url"`
	DurationMs       int     `json:"duration_ms"`
	TimeSignature    int     `json:"time_signature"`
}

// AudioAnalysis represents detailed audio analysis
type AudioAnalysis struct {
	Meta     *AnalysisMeta     `json:"meta"`
	Track    *AnalysisTrack    `json:"track"`
	Bars     []AnalysisBar     `json:"bars"`
	Beats    []AnalysisBeat    `json:"beats"`
	Sections []AnalysisSection `json:"sections"`
	Segments []AnalysisSegment `json:"segments"`
	Tatums   []AnalysisTatum   `json:"tatums"`
}

// AnalysisMeta represents analysis metadata
type AnalysisMeta struct {
	AnalyzerVersion string  `json:"analyzer_version"`
	Platform        string  `json:"platform"`
	DetailedStatus  string  `json:"detailed_status"`
	StatusCode      int     `json:"status_code"`
	Timestamp       int64   `json:"timestamp"`
	AnalysisTime    float64 `json:"analysis_time"`
	InputProcess    string  `json:"input_process"`
}

// AnalysisTrack represents track analysis
type AnalysisTrack struct {
	NumSamples              int     `json:"num_samples"`
	Duration                float64 `json:"duration"`
	SampleMD5               string  `json:"sample_md5"`
	OffsetSeconds           int     `json:"offset_seconds"`
	WindowSeconds           int     `json:"window_seconds"`
	AnalysisSampleRate      int     `json:"analysis_sample_rate"`
	AnalysisChannels        int     `json:"analysis_channels"`
	EndOfFadeIn             float64 `json:"end_of_fade_in"`
	StartOfFadeOut          float64 `json:"start_of_fade_out"`
	Loudness                float64 `json:"loudness"`
	Tempo                   float64 `json:"tempo"`
	TempoConfidence         float64 `json:"tempo_confidence"`
	TimeSignature           int     `json:"time_signature"`
	TimeSignatureConfidence float64 `json:"time_signature_confidence"`
	Key                     int     `json:"key"`
	KeyConfidence           float64 `json:"key_confidence"`
	Mode                    int     `json:"mode"`
	ModeConfidence          float64 `json:"mode_confidence"`
	Codestring              string  `json:"codestring"`
	CodeVersion             float64 `json:"code_version"`
	Echoprintstring         string  `json:"echoprintstring"`
	EchoprintVersion        float64 `json:"echoprint_version"`
	Synchstring             string  `json:"synchstring"`
	SynchVersion            float64 `json:"synch_version"`
	Rhythmstring            string  `json:"rhythmstring"`
	RhythmVersion           float64 `json:"rhythm_version"`
}

// AnalysisBar represents a bar in audio analysis
type AnalysisBar struct {
	Start      float64 `json:"start"`
	Duration   float64 `json:"duration"`
	Confidence float64 `json:"confidence"`
}

// AnalysisBeat represents a beat in audio analysis
type AnalysisBeat struct {
	Start      float64 `json:"start"`
	Duration   float64 `json:"duration"`
	Confidence float64 `json:"confidence"`
}

// AnalysisSection represents a section in audio analysis
type AnalysisSection struct {
	Start                   float64 `json:"start"`
	Duration                float64 `json:"duration"`
	Confidence              float64 `json:"confidence"`
	Loudness                float64 `json:"loudness"`
	Tempo                   float64 `json:"tempo"`
	TempoConfidence         float64 `json:"tempo_confidence"`
	Key                     int     `json:"key"`
	KeyConfidence           float64 `json:"key_confidence"`
	Mode                    int     `json:"mode"`
	ModeConfidence          float64 `json:"mode_confidence"`
	TimeSignature           int     `json:"time_signature"`
	TimeSignatureConfidence float64 `json:"time_signature_confidence"`
}

// AnalysisSegment represents a segment in audio analysis
type AnalysisSegment struct {
	Start           float64   `json:"start"`
	Duration        float64   `json:"duration"`
	Confidence      float64   `json:"confidence"`
	LoudnessStart   float64   `json:"loudness_start"`
	LoudnessMax     float64   `json:"loudness_max"`
	LoudnessMaxTime float64   `json:"loudness_max_time"`
	LoudnessEnd     float64   `json:"loudness_end"`
	Pitches         []float64 `json:"pitches"`
	Timbre          []float64 `json:"timbre"`
}

// AnalysisTatum represents a tatum in audio analysis
type AnalysisTatum struct {
	Start      float64 `json:"start"`
	Duration   float64 `json:"duration"`
	Confidence float64 `json:"confidence"`
}

// RecommendationSeed represents a recommendation seed
type RecommendationSeed struct {
	AfterFilteringSize int    `json:"afterFilteringSize"`
	AfterRelinkingSize int    `json:"afterRelinkingSize"`
	Href               string `json:"href"`
	ID                 string `json:"id"`
	InitialPoolSize    int    `json:"initialPoolSize"`
	Type               string `json:"type"`
}

// RecommendationsResponse represents a recommendations response
type RecommendationsResponse struct {
	Seeds  []RecommendationSeed `json:"seeds"`
	Tracks []Track              `json:"tracks"`
}

// CurrentlyPlaying represents currently playing track/episode
type CurrentlyPlaying struct {
	Timestamp            int64       `json:"timestamp"`
	Context              *Context    `json:"context"`
	ProgressMs           int         `json:"progress_ms"`
	IsPlaying            bool        `json:"is_playing"`
	Item                 interface{} `json:"item"` // Can be Track or Episode
	CurrentlyPlayingType string      `json:"currently_playing_type"`
	Actions              *Actions    `json:"actions"`
}

// Context represents playback context
type Context struct {
	ExternalURLs *ExternalURLs `json:"external_urls"`
	Href         string        `json:"href"`
	Type         string        `json:"type"`
	URI          string        `json:"uri"`
}

// QueueResponse represents the user's playback queue
type QueueResponse struct {
	CurrentlyPlaying *CurrentlyPlaying `json:"currently_playing"`
	Queue            []QueueItem       `json:"queue"`
}

// QueueItem represents an item in the queue (can be Track or Episode)
// Use interface{} with type assertion to handle both types
// Callers should use type assertion: item.(*Track) or item.(*Episode)
type QueueItem interface{}

// Actions represents available actions
type Actions struct {
	InterruptingPlayback  bool `json:"interrupting_playback"`
	Pausing               bool `json:"pausing"`
	Resuming              bool `json:"resuming"`
	Seeking               bool `json:"seeking"`
	SkippingNext          bool `json:"skipping_next"`
	SkippingPrev          bool `json:"skipping_prev"`
	TogglingRepeatContext bool `json:"toggling_repeat_context"`
	TogglingShuffle       bool `json:"toggling_shuffle"`
	TogglingRepeatTrack   bool `json:"toggling_repeat_track"`
	TransferringPlayback  bool `json:"transferring_playback"`
}

// Device represents a playback device
type Device struct {
	ID               *string `json:"id"`
	IsActive         bool    `json:"is_active"`
	IsPrivateSession bool    `json:"is_private_session"`
	IsRestricted     bool    `json:"is_restricted"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	VolumePercent    *int    `json:"volume_percent"`
}

// PlaybackState represents playback state
type PlaybackState struct {
	Device               *Device     `json:"device"`
	RepeatState          string      `json:"repeat_state"`
	ShuffleState         bool        `json:"shuffle_state"`
	Context              *Context    `json:"context"`
	Timestamp            int64       `json:"timestamp"`
	ProgressMs           int         `json:"progress_ms"`
	IsPlaying            bool        `json:"is_playing"`
	Item                 interface{} `json:"item"` // Can be Track or Episode
	CurrentlyPlayingType string      `json:"currently_playing_type"`
	Actions              *Actions    `json:"actions"`
}

// Category represents a browse category
type Category struct {
	Href  string  `json:"href"`
	Icons []Image `json:"icons"`
	ID    string  `json:"id"`
	Name  string  `json:"name"`
}

// PlayHistoryItem represents a play history item
type PlayHistoryItem struct {
	Track    Track    `json:"track"`
	PlayedAt string   `json:"played_at"` // ISO 8601 timestamp
	Context  *Context `json:"context,omitempty"`
}
