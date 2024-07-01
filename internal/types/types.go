package types

import "io"

type Config struct {
	MigrationFolder  string
	ConnectionString string
}

type Migration struct {
	Name       string
	FileHandle io.Reader
}

type MigrationGroupSummary struct {
	GroupId        int    `db:"group_id"`
	Name           string `db:"name"`
	MigrationCount int    `db:"migration_count"`
}

type MigrationGroup struct {
	GroupId    int
	Name       string
	Migrations []Migration
}

type Repository interface {
	EnsureCreated() error
	SummarizeMigrations() ([]MigrationGroupSummary, error)
	GetMigrationsByGroup(groupId int) ([]string, error)
	ExecuteMigrations([]MigrationGroup) error
}
