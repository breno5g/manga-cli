// core/downloader.go
package core

import (
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
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outDir, manga, chapter)
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		return err
	}

	return d.driver.DownloadChapter(manga, chapter, outputPath)
}

// GetChapters obtém a lista de capítulos disponíveis
func (d *Downloader) GetChapters(manga string) ([]string, error) {
	return d.driver.GetChapters(manga)
}
