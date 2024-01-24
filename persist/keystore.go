package persist

import (
	"database/sql"
	"fmt"
	"storage-go/keystore"
)

// KeyStorePersist is an interface for persisting keys

// Ensure SqlLiteConnection implements KeyStorePersist
var _ keystore.KeyStorePersist = (*SqlLiteConnection)(nil)

// SaveKey saves a serialized key in the database
func (conn *SqlLiteConnection) SaveKey(serializedKey keystore.SerializedKey) error {
	_, err := conn.dbConn.Exec("INSERT INTO keystore (id, nonce, key, tag) VALUES (?, ?, ?, ?)",
		serializedKey.ID, serializedKey.Nonce, serializedKey.Key, serializedKey.Tag)
	return err
}

// GetKey retrieves a serialized key by its ID
func (conn *SqlLiteConnection) GetKey(keyID string) (*keystore.SerializedKey, error) {
	var sk keystore.SerializedKey
	err := conn.dbConn.QueryRow("SELECT id, nonce, key, tag FROM keystore WHERE id = ?",
		keyID).Scan(&sk.ID, &sk.Nonce, &sk.Key, &sk.Tag)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("query failed: %v", err)
	default:
		return &sk, nil
	}
}

// GetWithTag retrieves a serialized key by its tag
func (conn *SqlLiteConnection) GetWithTag(tag string) (*keystore.SerializedKey, error) {
	var sk keystore.SerializedKey
	err := conn.dbConn.QueryRow("SELECT id, nonce, key, tag FROM keystore WHERE tag = ?",
		tag).Scan(&sk.ID, &sk.Nonce, &sk.Key, &sk.Tag)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("query failed: %v", err)
	default:
		return &sk, nil
	}
}
