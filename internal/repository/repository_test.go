package repository

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// in memory db
const dbName = "file:testDb?mode=memory"

type cleanUp func()

func getDb(t *testing.T) (*sqlx.DB, cleanUp) {
	db, err := sqlx.Open("sqlite3", dbName)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestCreateDatabase(t *testing.T) {
	t.Run("create new schema", func(t *testing.T) {
		db, cleanup := getDb(t)
		defer cleanup()

		repo := NewRepository(db)

		err := repo.EnsureCreated()
		require.Nil(t, err)

		query := `SELECT name 
			FROM sqlite_master 
			WHERE type='table' 
				AND name IN ($1, $2)`

		tables := make([]string, 2)

		err = db.Select(&tables, query, migrationTables[0], migrationTables[1])
		require.Nil(t, err)

		assert.Len(t, tables, 2)
	})

	t.Run("with existing schema", func(t *testing.T) {
		db, cleanup := getDb(t)
		defer cleanup()
		repo := NewRepository(db)

		cmd1 := `CREATE TABLE mmigration_group (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL);`

		if _, err := db.Exec(cmd1); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		cmd2 := `CREATE TABLE mmigration (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			migration_group_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			executed_at TIMESTAMP NOT NULL);`

		if _, err := db.Exec(cmd2); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		err := repo.EnsureCreated()
		if err != nil {
			t.Fatalf("failed to run EnsureCreated with up-to-date schema: %v", err)
		}
	})

	t.Run("with inconsistent schema", func(t *testing.T) {
		db, cleanup := getDb(t)
		defer cleanup()
		repo := NewRepository(db)

		cmd := `CREATE TABLE mmigration_group (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL
		);`

		if _, err := db.Exec(cmd); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		err := repo.EnsureCreated()
		if err != nil {
			t.Fatalf("failed to run EnsureCreated with inconsistent schema: %v", err)
		}
	})
}
