package db

import (
	"ctb-cli/filesyetem"
	"database/sql"
	"errors"
	"fmt"
)

var _ filesyetem.Persist = (*SqlLiteConnection)(nil)

// SavePath saves a serialized key in the database
func (conn *SqlLiteConnection) SavePath(path string, key string) error {
	_, err := conn.dbConn.Exec(
		"INSERT INTO filesystem (id, path) VALUES (?, ?)",
		key, path,
	)
	return err
}

// GetPath retrieves a path id by path
func (conn *SqlLiteConnection) GetPath(path string) (string, error) {
	var id string
	row := conn.dbConn.QueryRow(
		"SELECT id FROM filesystem WHERE path = ?",
		path,
	)
	err := row.Scan(&id)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("path not found in database (path): %s", path)
	case err != nil:
		return "", fmt.Errorf("query failed: %v", err)
	default:
		return id, nil
	}
}
