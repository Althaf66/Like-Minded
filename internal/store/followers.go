package store

import (
	"context"
	"database/sql"
)

type FollowStore struct {
	db *sql.DB
}

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

func (s *FollowStore) Follow(ctx context.Context, followid, id int64) error {
	query := `INSERT INTO followers (user_id,follower_id) VALUES ($1,$2)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, followid, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *FollowStore) Unfollow(ctx context.Context, followid, id int64) error {
	query := `DELETE FROM followers WHERE user_id =$1 AND follower_id = $2`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, followid, id)
	return err
}
