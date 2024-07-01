package types

import "errors"

var (
	ErrOnlyDirsInMigrationsFolder = errors.New("only directories are allowed in migration folder")
	ErrOnlyFilesInGroupFolder     = errors.New("only sql files are allowed in group folders")
)
