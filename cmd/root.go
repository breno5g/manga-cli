// cmd/root.go
package cmd

import (
	"github.com/breno5g/manga-cli/interfaces"
	"github.com/spf13/cobra"
)

var availableDrivers map[string]interfaces.Driver

// RootCmd representa o comando base quando chamado sem subcomandos
var RootCmd = &cobra.Command{
	Use:   "manga-cli",
	Short: "CLI para download de mangás de diversos sites",
	Long: `Uma ferramenta de linha de comando para baixar capítulos de mangás 
de diversos sites, salvando as imagens localmente em uma estrutura organizada de pastas.`,
}

// SetAvailableDrivers configura os drivers disponíveis para uso nos comandos
func SetAvailableDrivers(drivers map[string]interfaces.Driver) {
	availableDrivers = drivers
}

func init() {
	// Aqui adicionamos os subcomandos quando necessário
	RootCmd.AddCommand(downloadCmd)
}
