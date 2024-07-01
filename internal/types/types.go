package types

import "io"

type Config struct {
	MigrationFolder  string
	ConnectionString string
}

type Unit struct {
	Name       string
	FileHandle io.Reader
}

type GroupSummary struct {
	GroupId        int    `db:"group_id"`
	Name           string `db:"name"`
	MigrationCount int    `db:"migration_count"`
}

type Group struct {
	GroupId int
	Name    string
	Units   []Unit
}

type Repository interface {
	EnsureCreated() error
	SummarizeMigrations() ([]GroupSummary, error)
	GetMigrationsByGroup(groupId int) ([]string, error)
	ExecuteMigrations([]Group) error
}
