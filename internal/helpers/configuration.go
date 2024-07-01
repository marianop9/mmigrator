package helpers

import (
	"encoding/json"
	"os"

	"github.com/marianop9/mmigrator/internal/types"
)

func GetConfiguration(path string) (types.Config, error) {
	config := types.Config{}

	content, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(content, &config)

	return config, err
}
