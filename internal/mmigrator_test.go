package internal

import (
	"testing"

	"github.com/marianop9/mmigrator/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanMigrationFolder(t *testing.T) {
	t.Run("should return a succesfull folder scan", func(t *testing.T) {
		scan, err := scanMigrationsFolder("../test/migrationsFolder")

		require.Nil(t, err)

		require.NotNil(t, scan)

		group1Files, ok := scan["group1"]
		if assert.True(t, ok) {
			assert.Equal(t, len(group1Files), 1)
		}

		group2Files, ok := scan["group2"]
		if assert.True(t, ok) {
			assert.Equal(t, len(group2Files), 2)
		}

	})

	t.Run("should return an error because the migrations folder contains a file", func(t *testing.T) {
		_, err := scanMigrationsFolder("../test/invalidMigrationFolder1")

		assert.NotNil(t, err)

		assert.ErrorIs(t, err, types.ErrOnlyDirsInMigrationsFolder)
	})

	t.Run("should return an error because a group folder contains a subfolder", func(t *testing.T) {
		_, err := scanMigrationsFolder("../test/invalidMigrationFolder2")

		assert.NotNil(t, err)

		assert.ErrorIs(t, err, types.ErrOnlyFilesInGroupFolder)
	})
}

func TestCompareMigrations(t *testing.T) {
	t.Run("no migrations to update", func(t *testing.T) {
		scan, err := scanMigrationsFolder("../test/migrationsFolder")
		require.Nil(t, err)

		oldGroups := []types.GroupSummary{
			{
				Name:           "group1",
				MigrationCount: 1,
			},
			{
				Name:           "group2",
				MigrationCount: 2,
			},
		}

		groupsToUpdate := compareMigrations(oldGroups, scan)
		assert.Empty(t, groupsToUpdate)
	})

	t.Run("new group", func(t *testing.T) {
		scan, err := scanMigrationsFolder("../test/migrationsFolder")
		require.Nil(t, err)

		oldGroups := []types.GroupSummary{
			{
				Name:           "group1",
				MigrationCount: 1,
			},
		}

		groupsToUpdate := compareMigrations(oldGroups, scan)
		assert.Len(t, groupsToUpdate, 1)

		files := groupsToUpdate["group2"]
		if assert.NotNil(t, files) {
			assert.Len(t, files, 2)
		}
	})

	t.Run("new group and new file in existing group", func(t *testing.T) {
		scan, err := scanMigrationsFolder("../test/migrationsFolder")
		require.Nil(t, err)

		oldGroups := []types.GroupSummary{
			{
				Name:           "group2",
				MigrationCount: 1,
			},
		}

		groupsToUpdate := compareMigrations(oldGroups, scan)
		assert.Len(t, groupsToUpdate, 2)

		group1Files := groupsToUpdate["group1"]
		if assert.NotNil(t, group1Files) {
			assert.Len(t, group1Files, 1)
		}

		group2Files := groupsToUpdate["group2"]
		if assert.NotNil(t, group2Files) {
			assert.Len(t, group2Files, 2)
		}
	})
}

func TestGetExecutionUnits(t *testing.T) {
	migrationsFolder := "../test/migrationsFolder"

	t.Run("get units from brand new group", func(t *testing.T) {
		old := []string{}
		new := []string{
			"mig1.sql",
			"mig2.sql",
		}
		units, err := getExecutionUnits(migrationsFolder, "group2", old, new)
		require.Nil(t, err)

		assert.Len(t, units, 2)
	})

	t.Run("get units from existing new group", func(t *testing.T) {
		old := []string{
			"mig1.sql",
		}
		new := []string{
			"mig1.sql",
			"mig2.sql",
		}
		units, err := getExecutionUnits(migrationsFolder, "group2", old, new)
		require.Nil(t, err)

		assert.Len(t, units, 1)
	})
}
