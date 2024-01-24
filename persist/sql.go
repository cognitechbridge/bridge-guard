package persist

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

type SqlLiteConnection struct {
	dbConn *sql.DB
}

// NewSqlLiteConnection creates a new SqlLiteConnection
func NewSqlLiteConnection() (*SqlLiteConnection, error) {
	path, err := getUserPath()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(path, "db.db3")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &SqlLiteConnection{
		dbConn: conn,
	}, nil
}

// getUserPath returns the path to the .ctb directory in the user's home folder
func getUserPath() (string, error) {
	// Get the current user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Append the .ctb directory to the home directory
	ctbPath := filepath.Join(homeDir, ".ctb")

	// Optional: Create the .ctb directory if it doesn't exist
	err = os.MkdirAll(ctbPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return ctbPath, nil
}
