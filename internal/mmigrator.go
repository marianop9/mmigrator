package internal

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/marianop9/mmigrator/internal/types"
)

type Mmigrator struct {
	repo            types.Repository
	migrationFolder string
}

func NewMmigrator(repo types.Repository, migrrationFolder string) Mmigrator {
	return Mmigrator{
		repo,
		migrrationFolder,
	}
}

func (mm Mmigrator) Update(ctx context.Context) error {
	if err := mm.repo.EnsureCreated(); err != nil {
		return err
	}

	newGroups, err := scanMigrationsFolder(mm.migrationFolder)
	if err != nil {
		return err
	}

	oldGroups, err := mm.repo.SummarizeMigrations()
	if err != nil {
		return fmt.Errorf("failed to retrieve migration summary: %w", err)
	}

	groupsToUpdate := getGroupsToUpdate(oldGroups, newGroups)
	if len(groupsToUpdate) == 0 {
		fmt.Println("all migrations are up to date!")
	}

	// list of groups to execute
	// contain only migrations that have not been run yet
	groupsToExecute := make([]types.MigrationGroup, len(groupsToUpdate))

	for i, group := range groupsToUpdate {
		existingMigrationNames, err := mm.repo.GetMigrationsByGroup(group.GroupId)
		if err != nil {
			return fmt.Errorf("failed to retrieve migration names: %w", err)
		}

		executionGroup := types.MigrationGroup{
			GroupId:    group.GroupId,
			Name:       group.Name,
			Migrations: make([]types.Migration, 0),
		}

		// look for migrations not tracked by the previous version
		for _, migration := range group.Migrations {
			if sliceContains(existingMigrationNames, migration.Name) {
				continue
			}

			// get handle to new migration file
			filePath := path.Join(mm.migrationFolder, group.Name, migration.Name)

			migration.FileHandle, err = os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open '%s': %w", filePath, err)
			}

			// add migration to execution group
			executionGroup.Migrations = append(executionGroup.Migrations, migration)
		}

		// add execution group to list
		groupsToExecute[i] = executionGroup
	}

	mm.repo.ExecuteMigrations(groupsToExecute)

	return nil
}

func scanMigrationsFolder(basePath string) ([]types.MigrationGroup, error) {
	migrationDirs, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	migrationGroups := make([]types.MigrationGroup, len(migrationDirs))
	for i, dir := range migrationDirs {
		if !dir.IsDir() {
			return nil, types.ErrOnlyDirsInMigrationsFolder
		}

		migrationGroups[i] = types.MigrationGroup{
			GroupId:    0,
			Name:       dir.Name(),
			Migrations: make([]types.Migration, 0),
		}
	}

	for i := range migrationGroups {
		groupName := migrationGroups[i].Name
		groupPath := path.Join(basePath, groupName)

		entries, err := os.ReadDir(groupPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
				return nil, fmt.Errorf("%s: %v", groupName, types.ErrOnlyFilesInGroupFolder)
			}
			migration := types.Migration{
				Name:       entry.Name(),
				FileHandle: nil,
			}
			migrationGroups[i].Migrations = append(migrationGroups[i].Migrations, migration)
		}
	}

	return migrationGroups, nil
}

func getGroupsToUpdate(
	oldGroups []types.MigrationGroupSummary,
	newGroups []types.MigrationGroup,
) []types.MigrationGroup {
	groupsToUpdate := make([]types.MigrationGroup, 0)
	// compare migration groups from folder and db
	for _, old := range oldGroups {
		new := findByName(newGroups, old.Name)
		if new == nil {
			fmt.Printf("WARNING: previously executed group %s not found in migrations folder\n", old.Name)
			continue
		}

		if old.MigrationCount < len(new.Migrations) {
			// this group has new files
			// save the groupId and add to list
			new.GroupId = old.GroupId
			groupsToUpdate = append(groupsToUpdate, *new)
		} else if old.MigrationCount > len(new.Migrations) {
			// the current version of the group has less files than the one saved on db
			fmt.Printf("WARNING: group '%s' has less migrations than the previous version\n", old.Name)
		}
	}

	return groupsToUpdate
}

func findByName(groups []types.MigrationGroup, name string) *types.MigrationGroup {
	for _, group := range groups {
		if group.Name == name {
			return &group
		}
	}

	return nil
}

func sliceContains[T comparable](slice []T, target T) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}
