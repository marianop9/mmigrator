package helpers

import (
	"encoding/json"
	"os"
	"path"

	"github.com/marianop9/mmigrator/internal/types"
)

const mmigratorConfigFile = "mmigrator-config.json"

func GetConfiguration(configPath string) (types.Config, error) {
	config := types.Config{}

	file := path.Join(configPath, mmigratorConfigFile)

	content, err := os.ReadFile(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(content, &config)

	return config, err
}
