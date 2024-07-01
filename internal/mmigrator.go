package internal

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/marianop9/mmigrator/internal/types"
)

type folderScan map[string][]string

type Mmigrator struct {
	migrationsFolder string
	repo             types.Repository
}

func New(repo types.Repository, migrationsFolder string) Mmigrator {
	return Mmigrator{
		repo:             repo,
		migrationsFolder: migrationsFolder,
	}
}

func (mm Mmigrator) Update(ctx context.Context) error {
	var err error

	if err = mm.repo.EnsureCreated(); err != nil {
		return err
	}

	fmt.Printf("scanning migration folder: %s\n", mm.migrationsFolder)
	folderScan, err := scanMigrationsFolder(mm.migrationsFolder)
	if err != nil {
		return err
	}

	fmt.Println("retrieving previous migrations...")
	oldGroups, err := mm.repo.SummarizeMigrations()
	if err != nil {
		return fmt.Errorf("failed to retrieve migration summary: %w", err)
	}

	// get groups that need to be updated
	groupsToUpdate := compareMigrations(oldGroups, folderScan)
	if len(groupsToUpdate) == 0 {
		fmt.Println("all migrations are up to date")
		return nil
	}
	// fmt.Printf("groupsToUpdate: %#v\n", groupsToUpdate)

	// determine migrations not yet run
	groupsToExecute, err := mm.getExecutionGroups(oldGroups, groupsToUpdate)
	if err != nil {
		return err
	}

	return mm.repo.ExecuteMigrations(groupsToExecute)
}

func scanMigrationsFolder(basePath string) (map[string][]string, error) {
	migrationDirs, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	migrationGroups := make(map[string][]string, len(migrationDirs))
	for _, dir := range migrationDirs {
		if !dir.IsDir() {
			return nil, types.ErrOnlyDirsInMigrationsFolder
		}

		migrationGroups[dir.Name()] = []string{}
	}

	for groupName := range migrationGroups {
		groupPath := path.Join(basePath, groupName)

		entries, err := os.ReadDir(groupPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
				return nil, fmt.Errorf("%s: %w", groupName, types.ErrOnlyFilesInGroupFolder)
			}
			migrationGroups[groupName] = append(migrationGroups[groupName], entry.Name())
		}
	}

	return migrationGroups, nil
}

func compareMigrations(oldGroups []types.GroupSummary, folderScan folderScan) folderScan {
	// compare migration groups from folder and db
	for _, old := range oldGroups {
		files, ok := folderScan[old.Name]
		if !ok {
			fmt.Printf("WARNING: previously executed group %s not found in migrations folder\n", old.Name)
			continue
		}

		if old.MigrationCount == len(files) {
			delete(folderScan, old.Name)
		} else if old.MigrationCount > len(files) {
			// the current version of the group has less files than the one saved on db
			fmt.Printf("WARNING: group '%s' has less migrations than the previous version\n", old.Name)
			delete(folderScan, old.Name)
		}
	}

	return folderScan
}

func (mm Mmigrator) getExecutionGroups(oldGroups []types.GroupSummary, groupsToUpdate folderScan) ([]types.Group, error) {
	executionGroups := make([]types.Group, 0, len(groupsToUpdate))

	for group, units := range groupsToUpdate {
		oldGroup := findByName(oldGroups, group)

		groupId := 0
		oldUnits := []string{}

		if oldGroup != nil {
			groupId = oldGroup.GroupId

			var err error
			if oldUnits, err = mm.repo.GetMigrationsByGroup(groupId); err != nil {
				return nil, err
			}
		}

		executionUnits, err := getExecutionUnits(mm.migrationsFolder, group, oldUnits, units)
		if err != nil {
			return nil, err
		}

		group := types.Group{
			GroupId: groupId,
			Name:    group,
			Units:   executionUnits,
		}
		executionGroups = append(executionGroups, group)
	}

	return executionGroups, nil
}

func getExecutionUnits(migrationsFolder, groupName string, oldMigrations []string, newMigrations []string) ([]types.Unit, error) {
	units := make([]types.Unit, 0)

	// look for migrations not tracked by the previous version
	for _, migration := range newMigrations {
		if sliceContains(oldMigrations, migration) {
			continue
		}

		// get handle to new migration file
		filePath := path.Join(migrationsFolder, groupName, migration)

		fileHandle, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open '%s': %w", filePath, err)
		}

		units = append(units, types.Unit{
			Name:       migration,
			FileHandle: fileHandle,
		})
	}
	return units, nil
}

func findByName(groups []types.GroupSummary, name string) *types.GroupSummary {
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
