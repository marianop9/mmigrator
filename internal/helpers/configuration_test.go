package helpers

import "testing"

func Test(t *testing.T) {
	path := "../../test"

	config, err := GetConfiguration(path)

	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}

	expectedConnString := "test.db"
	expectedMigFolder := "test/migrationsFolder"
	if config.ConnectionString != expectedConnString {
		t.Fatalf("connectionString: expected %s, got %s", expectedConnString, config.ConnectionString)
	}
	if config.MigrationFolder != expectedMigFolder {
		t.Fatalf("migrationFolder: expected %s, got %s", expectedMigFolder, config.MigrationFolder)
	}
}
