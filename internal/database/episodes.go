package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Episode represents a YouTube video in the database
type Episode struct {
	ID          int       `json:"id"`
	VideoID     string    `json:"video_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PublishedAt time.Time `json:"published_at"`
	Image1      string    `json:"image1"` // Default thumbnail (120x90)
	Image2      string    `json:"image2"` // Medium thumbnail (320x180)
	Image3      string    `json:"image3"` // High res thumbnail (480x360)
	ViewCount   int       `json:"view_count"`
	VideoSite   int       `json:"video_site"` // 1 = YouTube
	Active      bool      `json:"active"`
	SortOrder   *int      `json:"sort_order,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpsertEpisode inserts or updates an episode in the database (idempotent operation)
func (c *Client) UpsertEpisode(ctx context.Context, episode *Episode) error {
	query := `
		INSERT INTO episodes (
			video_id, title, description, published_at,
			image1, image2, image3, view_count,
			video_site, active, sort_order, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (video_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			published_at = EXCLUDED.published_at,
			image1 = EXCLUDED.image1,
			image2 = EXCLUDED.image2,
			image3 = EXCLUDED.image3,
			view_count = EXCLUDED.view_count,
			active = EXCLUDED.active,
			sort_order = EXCLUDED.sort_order,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := c.db.QueryRowContext(
		ctx,
		query,
		episode.VideoID,
		episode.Title,
		episode.Description,
		episode.PublishedAt,
		episode.Image1,
		episode.Image2,
		episode.Image3,
		episode.ViewCount,
		episode.VideoSite,
		episode.Active,
		episode.SortOrder,
	).Scan(&episode.ID, &episode.CreatedAt, &episode.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert episode: %w", err)
	}

	return nil
}

// GetEpisodeByVideoID retrieves an episode by YouTube video ID
func (c *Client) GetEpisodeByVideoID(ctx context.Context, videoID string) (*Episode, error) {
	query := `
		SELECT id, video_id, title, description, published_at,
			image1, image2, image3, view_count, video_site,
			active, sort_order, created_at, updated_at
		FROM episodes
		WHERE video_id = $1
	`

	episode := &Episode{}
	err := c.db.QueryRowContext(ctx, query, videoID).Scan(
		&episode.ID,
		&episode.VideoID,
		&episode.Title,
		&episode.Description,
		&episode.PublishedAt,
		&episode.Image1,
		&episode.Image2,
		&episode.Image3,
		&episode.ViewCount,
		&episode.VideoSite,
		&episode.Active,
		&episode.SortOrder,
		&episode.CreatedAt,
		&episode.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("episode not found: %s", videoID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get episode: %w", err)
	}

	return episode, nil
}

// GetEpisodeByID retrieves an episode by internal database ID
func (c *Client) GetEpisodeByID(ctx context.Context, id int) (*Episode, error) {
	query := `
		SELECT id, video_id, title, description, published_at,
			image1, image2, image3, view_count, video_site,
			active, sort_order, created_at, updated_at
		FROM episodes
		WHERE id = $1
	`

	episode := &Episode{}
	err := c.db.QueryRowContext(ctx, query, id).Scan(
		&episode.ID,
		&episode.VideoID,
		&episode.Title,
		&episode.Description,
		&episode.PublishedAt,
		&episode.Image1,
		&episode.Image2,
		&episode.Image3,
		&episode.ViewCount,
		&episode.VideoSite,
		&episode.Active,
		&episode.SortOrder,
		&episode.CreatedAt,
		&episode.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("episode not found: id=%d", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get episode: %w", err)
	}

	return episode, nil
}

// MarkEpisodeInactive marks an episode as inactive (for deleted/private videos)
func (c *Client) MarkEpisodeInactive(ctx context.Context, videoID string) error {
	query := `
		UPDATE episodes
		SET active = false, updated_at = NOW()
		WHERE video_id = $1
	`

	result, err := c.db.ExecContext(ctx, query, videoID)
	if err != nil {
		return fmt.Errorf("failed to mark episode inactive: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("episode not found: %s", videoID)
	}

	return nil
}

// ListEpisodes retrieves episodes with pagination
func (c *Client) ListEpisodes(ctx context.Context, limit, offset int) ([]*Episode, error) {
	query := `
		SELECT id, video_id, title, description, published_at,
			image1, image2, image3, view_count, video_site,
			active, sort_order, created_at, updated_at
		FROM episodes
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := c.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list episodes: %w", err)
	}
	defer rows.Close()

	var episodes []*Episode
	for rows.Next() {
		episode := &Episode{}
		err := rows.Scan(
			&episode.ID,
			&episode.VideoID,
			&episode.Title,
			&episode.Description,
			&episode.PublishedAt,
			&episode.Image1,
			&episode.Image2,
			&episode.Image3,
			&episode.ViewCount,
			&episode.VideoSite,
			&episode.Active,
			&episode.SortOrder,
			&episode.CreatedAt,
			&episode.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}
		episodes = append(episodes, episode)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating episodes: %w", err)
	}

	return episodes, nil
}
