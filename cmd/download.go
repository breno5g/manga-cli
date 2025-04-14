// cmd/download.go
package cmd

import (
	"fmt"
	"os"

	"github.com/breno5g/manga-cli/core"
	"github.com/breno5g/manga-cli/drivers/mangadex"
	"github.com/breno5g/manga-cli/interfaces"
	"github.com/spf13/cobra"
)

var (
	site     string
	manga    string
	chapter  string
	language string
	outDir   string
)

// downloadCmd representa o comando para baixar capítulos de mangá
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Baixa capítulos de mangá",
	Long:  "Baixa capítulos de mangá de sites suportados",
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedDriver interfaces.Driver
		for _, d := range availableDrivers {
			if driver, ok := d.(interfaces.Driver); ok {
				selectedDriver = driver
				break
			}
		}

		if selectedDriver == nil {
			return fmt.Errorf("driver não encontrado para o site: %s", site)
		}

		if mangaDexDriver, ok := selectedDriver.(*mangadex.MangaDexDriver); ok {
			if langDriver, ok := mangaDexDriver.(interfaces.LanguageSupportDriver); ok {
				langDriver.SetLanguage(language)
			}
		}

		downloader := core.NewDownloader(selectedDriver)
		return downloader.DownloadChapter(manga, chapter, outDir)
	},
}

func init() {
	// Definir flags para o comando download
	downloadCmd.Flags().StringVarP(&site, "site", "s", "", "Site de origem do mangá")
	downloadCmd.Flags().StringVarP(&manga, "manga", "m", "", "Nome do mangá")
	downloadCmd.Flags().StringVarP(&chapter, "chapter", "c", "", "Número do capítulo")
	downloadCmd.Flags().StringVarP(&language, "language", "l", "pt-br", "Idioma das traduções (apenas para MangaDex)")
	downloadCmd.Flags().StringVarP(&outDir, "output", "o", "downloads", "Diretório de saída")

	// Marcar flags obrigatórias
	downloadCmd.MarkFlagRequired("site")
	downloadCmd.MarkFlagRequired("manga")
	downloadCmd.MarkFlagRequired("chapter")
}
