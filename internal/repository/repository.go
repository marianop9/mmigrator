package repository

import (
	"database/sql"
	"fmt"
	"io"
	"time"

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
		name TEXT NOT NULL
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
		name TEXT NOT NULL,
		executed_at TEXT NOT NULL
	);`

	if _, sqlErr := tx.Exec(cmd); sqlErr != nil {
		return sqlErr
	}

	return nil
}

func (r Repository) SummarizeMigrations() ([]types.GroupSummary, error) {
	sql := `SELECT mg.id "group_id",
				mg.name "name",
				count(*) "migration_count"
		FROM mmigration_group mg 
			JOIN mmigration m ON m.migration_group_id = mg.id
		GROUP BY mg.id;
	`

	summaries := []types.GroupSummary{}

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

func (r Repository) ExecuteMigrations(groups []types.Group) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fmt.Println("EXECUTING MIGRATIONS")

	for _, group := range groups {
		fmt.Printf("\t executing group '%s':\n", group.Name)

		if err := applyMigrationgGroup(tx, group); err != nil {
			return fmt.Errorf("failed to execute group '%s', %v", group.Name, err)
		}

		if err := logMigrationGroup(tx, group); err != nil {
			return fmt.Errorf("failed to log group '%s', %v", group.Name, err)
		}

		fmt.Printf("\t done executing group '%s'\n", group.Name)
	}

	return tx.Commit()
}

func applyMigrationgGroup(tx *sql.Tx, group types.Group) error {
	for _, unit := range group.Units {
		buf, err := io.ReadAll(unit.FileHandle)

		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(buf)); err != nil {
			return fmt.Errorf("failed to execute %s: %w", group.Name, err)
		}

		fmt.Printf("\t\t * executed %s\n", unit.Name)
	}

	return nil
}

func logMigrationGroup(tx *sql.Tx, group types.Group) error {
	if group.GroupId == 0 {
		groupCmd := `INSERT INTO mmigration_group (name) VALUES ($1)`

		result, err := tx.Exec(groupCmd, group.Name)
		if err != nil {
			return err
		}

		groupId, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to retrieve lastInsertId: %w", err)
		}
		group.GroupId = int(groupId)
	}

	unitCmd := `INSERT INTO mmigration (
		migration_group_id,
		name,
		executed_at
	) VALUES ($1, $2, $3);`

	now := time.Now().Format(time.RFC822)

	for _, unit := range group.Units {
		_, err := tx.Exec(unitCmd, group.GroupId, unit.Name, now)
		if err != nil {
			return err
		}
	}

	return nil
}
