package store

import (
	"context"
	"database/sql"
)

type RoleStore struct {
	db *sql.DB
}

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

func (r *RoleStore) GetByName(ctx context.Context, name string) (*Role, error) {
	query := "SELECT id,name,description,level FROM roles WHERE name = $1"

	role := &Role{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description, &role.Level)
	if err != nil {
		return nil, err
	}
	return role, nil
}
