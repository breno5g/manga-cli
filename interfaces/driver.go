// interfaces/driver.go
package interfaces

// Driver define a interface comum que todos os drivers de sites de mangá devem implementar
type Driver interface {
	// GetChapters retorna a lista de capítulos disponíveis para um mangá específico
	GetChapters(mangaName string) ([]string, error)

	// DownloadChapter baixa um capítulo específico de um mangá e salva no diretório de saída
	DownloadChapter(mangaName, chapterNumber, outputDir string) error
}

// LanguageSupportDriver é uma interface adicional para drivers que suportam múltiplos idiomas
type LanguageSupportDriver interface {
	Driver

	// SetLanguage define o idioma das traduções
	SetLanguage(language string)
}
