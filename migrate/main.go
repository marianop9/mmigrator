package main

import (
	"fmt"
	"os"

	"github.com/marianop9/mmigrator"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("missing config path")
		return
	}

	configPath := os.Args[1]
	mmigrator.Run(configPath)
}
