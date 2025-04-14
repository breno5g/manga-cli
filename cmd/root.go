// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "manga-cli",
	Short: "CLI para download de mangás",
	Long:  "Uma ferramenta de linha de comando para baixar mangás de vários sites",
}

var availableDrivers []interface{}

func SetAvailableDrivers(drivers []interface{}) {
	availableDrivers = drivers
}

func init() {
	RootCmd.AddCommand(downloadCmd)
}

func Execute() error {
	return RootCmd.Execute()
}
