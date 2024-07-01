package mmigrator

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/marianop9/mmigrator/internal"
	"github.com/marianop9/mmigrator/internal/helpers"
	"github.com/marianop9/mmigrator/internal/repository"
)

func Run(configPath string) {
	config, err := helpers.GetConfiguration(configPath)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	db := sqlx.MustOpen("sqlite3", config.ConnectionString)

	repo := repository.NewRepository(db)

	mmigrator := internal.New(repo, config.MigrationFolder)

	if err := mmigrator.Update(context.Background()); err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}
