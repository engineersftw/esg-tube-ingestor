package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Playlist represents a YouTube playlist in the database
type Playlist struct {
	ID          int64
	PlaylistID  string
	Title       string
	Description string
	Slug        string
	Image1      string
	Image2      string
	Image3      string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PlaylistItem represents a video in a playlist
type PlaylistItem struct {
	ID         int64
	PlaylistID int64 // FK to playlists.id
	EpisodeID  int64 // FK to episodes.id
	SortOrder  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// UpsertPlaylist inserts or updates a playlist in the database
func (c *Client) UpsertPlaylist(ctx context.Context, playlist *Playlist) error {
	query := `
		INSERT INTO playlists (playlist_id, title, description, slug, image1, image2, image3, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (playlist_id)
		DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			slug = EXCLUDED.slug,
			image1 = EXCLUDED.image1,
			image2 = EXCLUDED.image2,
			image3 = EXCLUDED.image3,
			active = EXCLUDED.active,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := c.db.QueryRowContext(
		ctx,
		query,
		playlist.PlaylistID,
		playlist.Title,
		playlist.Description,
		playlist.Slug,
		playlist.Image1,
		playlist.Image2,
		playlist.Image3,
		playlist.Active,
		now,
		now,
	).Scan(&playlist.ID, &playlist.CreatedAt, &playlist.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert playlist: %w", err)
	}

	return nil
}

// GetPlaylistByPlaylistID retrieves a playlist by its YouTube playlist ID
func (c *Client) GetPlaylistByPlaylistID(ctx context.Context, playlistID string) (*Playlist, error) {
	query := `
		SELECT id, playlist_id, title, description, slug, image1, image2, image3, active, created_at, updated_at
		FROM playlists
		WHERE playlist_id = $1
	`

	playlist := &Playlist{}
	err := c.db.QueryRowContext(ctx, query, playlistID).Scan(
		&playlist.ID,
		&playlist.PlaylistID,
		&playlist.Title,
		&playlist.Description,
		&playlist.Slug,
		&playlist.Image1,
		&playlist.Image2,
		&playlist.Image3,
		&playlist.Active,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}

	return playlist, nil
}

// UpsertPlaylistItem inserts or updates a playlist item
func (c *Client) UpsertPlaylistItem(ctx context.Context, item *PlaylistItem) error {
	query := `
		INSERT INTO playlist_items (playlist_id, episode_id, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (playlist_id, episode_id)
		DO UPDATE SET
			sort_order = EXCLUDED.sort_order,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at
	`

	now := time.Time{}
	err := c.db.QueryRowContext(
		ctx,
		query,
		item.PlaylistID,
		item.EpisodeID,
		item.SortOrder,
		now,
		now,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert playlist item: %w", err)
	}

	return nil
}

// DeletePlaylistItems removes all items from a playlist
func (c *Client) DeletePlaylistItems(ctx context.Context, playlistID int64) error {
	query := `DELETE FROM playlist_items WHERE playlist_id = $1`

	_, err := c.db.ExecContext(ctx, query, playlistID)
	if err != nil {
		return fmt.Errorf("failed to delete playlist items: %w", err)
	}

	return nil
}
