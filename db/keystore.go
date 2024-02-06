package db

import (
	"ctb-cli/keystore"
	"database/sql"
	"errors"
	"fmt"
)

// Ensure SqlLiteConnection implements KeyStorePersist
var _ keystore.Persist = (*SqlLiteConnection)(nil)

// SaveKey saves a serialized key in the database
func (conn *SqlLiteConnection) SaveKey(serializedKey keystore.SerializedKey) error {
	_, err := conn.dbConn.Exec(
		"INSERT INTO keystore (id, nonce, key, tag) VALUES (?, ?, ?, NULL)",
		serializedKey.ID, serializedKey.Nonce, serializedKey.Key,
	)
	return err
}

// GetKey retrieves a serialized key by its ID
func (conn *SqlLiteConnection) GetKey(keyID string) (*keystore.SerializedKey, error) {
	var sk keystore.SerializedKey
	row := conn.dbConn.QueryRow(
		"SELECT id, nonce, key FROM keystore WHERE id = ?",
		keyID,
	)
	err := row.Scan(&sk.ID, &sk.Nonce, &sk.Key)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("key not found in database (key): %s", keyID)
	case err != nil:
		return nil, fmt.Errorf("query failed: %v", err)
	default:
		return &sk, nil
	}
}
