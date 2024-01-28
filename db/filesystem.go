package db

import (
	"ctb-cli/filesyetem"
	"database/sql"
	"errors"
	"fmt"
)

var _ filesyetem.Persist = (*SqlLiteConnection)(nil)

// SavePath saves a serialized key in the database
func (conn *SqlLiteConnection) SavePath(path string, key string, isDir bool) error {
	isDirN := 0
	if isDir == true {
		isDirN = 1
	}
	_, err := conn.dbConn.Exec(
		"INSERT INTO filesystem (id, path, isDir) VALUES (?, ?, ?)",
		key, path, isDirN,
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

func (conn *SqlLiteConnection) RemovePath(path string) error {
	_, err := conn.dbConn.Exec(
		"DELETE FROM filesystem WHERE path = ?",
		path,
	)
	return err
}

func (conn *SqlLiteConnection) PathExist(path string) (bool, error) {
	var id string
	row := conn.dbConn.QueryRow(
		"SELECT id FROM filesystem WHERE path = ? OR path LIKE ?",
		path, path+"/%",
	)
	err := row.Scan(&id)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("query failed: %v", err)
	default:
		return true, nil
	}
}
