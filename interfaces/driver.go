// interfaces/driver.go
package interfaces

// Driver define a interface comum que todos os drivers de sites de mangá devem implementar
type Driver interface {
	// GetChapters retorna a lista de capítulos disponíveis para um mangá específico
	GetChapters(manga string) ([]string, error)

	// DownloadChapter baixa um capítulo específico de um mangá e salva no diretório de saída
	DownloadChapter(manga string, chapter string, outDir string) error
}
