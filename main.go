// main.go
package main

import (
	"github.com/breno5g/manga-cli/cmd"
	"github.com/breno5g/manga-cli/drivers/mangadex"
	"github.com/breno5g/manga-cli/drivers/unionmangas"
)

func main() {
	drivers := []interface{}{
		mangadex.NewDriver(),
		unionmangas.NewDriver(),
	}

	cmd.SetAvailableDrivers(drivers)
	cmd.Execute()
}
