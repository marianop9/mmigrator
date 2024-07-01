package repository

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/marianop9/mmigrator/internal/types"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return Repository{db}
}

var migrationTables = [2]string{
	"mmigration_group",
	"mmigration",
}

func (r Repository) EnsureCreated() error {
	query := `SELECT name 
		FROM sqlite_master 
		WHERE type='table' 
			AND name IN ($1, $2)`

	tables := make([]string, 2)

	err := r.db.Select(&tables, query, migrationTables[0], migrationTables[1])
	if err != nil {
		return err
	}

	switch len(tables) {
	case 2:
		fmt.Println("migrations tables exist")
		return nil
	case 1:
		fmt.Println("inconsistent schema - recreating tables...")

		if _, err := r.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", migrationTables[0])); err != nil {
			return fmt.Errorf("inconsistent schema - failed to drop table: %w", err)
		}
		if _, err := r.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", migrationTables[1])); err != nil {
			return fmt.Errorf("inconsistent schema - failed to drop table: %w", err)
		}
		fmt.Println("dropped migration tables...")

		fallthrough
	case 0:
		fmt.Println("creating migration tables...")

		tx, err := r.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start tx: %w", err)
		}
		defer tx.Rollback()

		if err = createMigrationGroupTable(tx); err != nil {
			return err
		}

		if err = createMigrationTable(tx); err != nil {
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}
	default:
		panic(fmt.Sprintf("got %d tables, but expected 2 at most", len(migrationTables)))
	}

	return nil
}

func createMigrationGroupTable(tx *sql.Tx) error {
	fmt.Println("creating table 'mmigration_group'...")

	cmd := `CREATE TABLE mmigration_group (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL
	);`

	if _, sqlErr := tx.Exec(cmd); sqlErr != nil {
		return sqlErr
	}

	return nil
}

func createMigrationTable(tx *sql.Tx) error {
	fmt.Println(`creating table 'mmigration'...`)

	cmd := `CREATE TABLE mmigration (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		migration_group_id INTEGER NOT NULL,
		name VARCHAR(255) NOT NULL,
		executed_at TIMESTAMP NOT NULL
	);`

	if _, sqlErr := tx.Exec(cmd); sqlErr != nil {
		return sqlErr
	}

	return nil
}

func (r Repository) SummarizeMigrations() ([]types.MigrationGroupSummary, error) {
	sql := `SELECT mg.id "group_id",
				mg.name "name",
				count(*) "migration_count"
		FROM mmigration_group mg 
			JOIN mmigration m ON m.migration_group_id = mg.id
		GROUP BY mg.id;
	`

	summaries := []types.MigrationGroupSummary{}

	err := r.db.Select(&summaries, sql)
	return summaries, err
}

func (r Repository) GetMigrationsByGroup(groupId int) ([]string, error) {
	sql := `SELECT m.name
		FROM mmigration m
		WHERE m.migration_group_id = $1
	`
	migrations := []string{}

	err := r.db.Select(&migrations, sql, groupId)
	return migrations, err
}
