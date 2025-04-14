// cmd/download.go
package cmd

import (
	"fmt"
	"strconv"

	"github.com/breno5g/manga-cli/core"
	"github.com/spf13/cobra"
)

var (
	site   string
	manga  string
	start  int
	end    int
	output string
)

// downloadCmd representa o comando para baixar capítulos de mangá
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Baixa capítulos de mangá",
	Long:  `Faz o download de um intervalo de capítulos de um mangá específico.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Verificar se o site é suportado
		driver, exists := availableDrivers[site]
		if !exists {
			return fmt.Errorf("site não suportado: %s", site)
		}

		// Criar downloader com o driver selecionado
		downloader := core.NewDownloader(driver)

		// Fazer download dos capítulos
		fmt.Printf("Baixando %s, capítulos %d a %d do site %s\n", manga, start, end, site)
		for i := start; i <= end; i++ {
			chapter := strconv.Itoa(i)
			fmt.Printf("Baixando capítulo %s...\n", chapter)
			if err := downloader.DownloadChapter(manga, chapter, output); err != nil {
				fmt.Printf("Erro ao baixar capítulo %s: %v\n", chapter, err)
				continue
			}
			fmt.Printf("Capítulo %s baixado com sucesso!\n", chapter)
		}

		return nil
	},
}

func init() {
	// Definir flags para o comando download
	downloadCmd.Flags().StringVar(&site, "site", "", "Site de origem do mangá (obrigatório)")
	downloadCmd.Flags().StringVar(&manga, "manga", "", "Nome do mangá (obrigatório)")
	downloadCmd.Flags().IntVar(&start, "start", 1, "Capítulo inicial")
	downloadCmd.Flags().IntVar(&end, "end", 1, "Capítulo final")
	downloadCmd.Flags().StringVar(&output, "output", "./downloads", "Diretório de saída")

	// Marcar flags obrigatórias
	downloadCmd.MarkFlagRequired("site")
	downloadCmd.MarkFlagRequired("manga")
}
