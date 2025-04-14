// main.go
package main

import (
	"fmt"
	"os"

	"github.com/breno5g/manga-cli/cmd"
	"github.com/breno5g/manga-cli/drivers/mangadex"
	"github.com/breno5g/manga-cli/drivers/unionmangas"
	"github.com/breno5g/manga-cli/interfaces"
)

// Registrar todos os drivers disponíveis
var availableDrivers = map[string]interfaces.Driver{
	"mangadex":    mangadex.NewDriver(),
	"unionmangas": unionmangas.NewDriver(),
}

func main() {
	cmd.SetAvailableDrivers(availableDrivers)
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
