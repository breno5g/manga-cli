// core/downloader.go
package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/breno5g/manga-cli/interfaces"
)

// Downloader é responsável por orquestrar o download de capítulos
type Downloader struct {
	driver interfaces.Driver
}

// NewDownloader cria uma nova instância de Downloader com o driver especificado
func NewDownloader(driver interfaces.Driver) *Downloader {
	return &Downloader{
		driver: driver,
	}
}

// DownloadChapter faz o download de um capítulo específico
func (d *Downloader) DownloadChapter(manga, chapter, outDir string) error {
	// Criar estrutura de diretórios
	chapterDir := filepath.Join(outDir, manga, chapter)
	if err := os.MkdirAll(chapterDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %w", err)
	}

	// Delegar o download para o driver específico
	return d.driver.DownloadChapter(manga, chapter, chapterDir)
}

// GetChapters obtém a lista de capítulos disponíveis
func (d *Downloader) GetChapters(manga string) ([]string, error) {
	return d.driver.GetChapters(manga)
}
