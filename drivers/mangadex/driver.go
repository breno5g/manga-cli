// drivers/mangadex/driver.go
package mangadex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/breno5g/manga-cli/interfaces"
)

// MangaDexDriver implementa a interface Driver para o site MangaDex
type MangaDexDriver struct {
	client   *http.Client
	language string
}

// NewDriver cria uma nova instância do driver MangaDex
func NewDriver() interfaces.Driver {
	return &MangaDexDriver{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		language: "pt-br", // Define português do Brasil como idioma padrão
	}
}

// API de busca do MangaDex (simplificada)
const (
	apiBaseURL      = "https://api.mangadex.org"
	searchEndpoint  = "/manga"
	chapterEndpoint = "/chapter"
)

// mangaSearchResult representa a resposta da API de busca do MangaDex
type mangaSearchResult struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Title map[string]string `json:"title"`
		} `json:"attributes"`
	} `json:"data"`
}

// chapterResult representa a resposta da API de capítulos do MangaDex
type chapterResult struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Chapter            string `json:"chapter"`
			TranslatedLanguage string `json:"translatedLanguage"`
		} `json:"attributes"`
	} `json:"data"`
}

// chapterPagesResult representa a resposta com os dados de páginas de um capítulo
type chapterPagesResult struct {
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash      string   `json:"hash"`
		Data      []string `json:"data"`
		DataSaver []string `json:"dataSaver"`
	} `json:"chapter"`
}

// GetChapters retorna a lista de capítulos disponíveis para o mangá
func (d *MangaDexDriver) GetChapters(manga string) ([]string, error) {
	// 1. Encontrar o ID do mangá pelo nome
	mangaID, err := d.findMangaID(manga)
	if err != nil {
		return nil, err
	}

	// 2. Buscar capítulos pelo ID do mangá
	req, err := http.NewRequest("GET", apiBaseURL+chapterEndpoint, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("manga", mangaID)
	q.Add("translatedLanguage[]", d.language) // Filtrar por idioma português
	q.Add("limit", "100")                     // Limite de resultados por página
	req.URL.RawQuery = q.Encode()

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao buscar capítulos: %s", resp.Status)
	}

	var result chapterResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Extrair números dos capítulos
	chapters := make([]string, 0, len(result.Data))
	for _, ch := range result.Data {
		if ch.Attributes.Chapter != "" {
			chapters = append(chapters, ch.Attributes.Chapter)
		}
	}

	return chapters, nil
}

// findMangaID busca o ID do mangá pelo nome
func (d *MangaDexDriver) findMangaID(mangaName string) (string, error) {
	req, err := http.NewRequest("GET", apiBaseURL+searchEndpoint, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("title", mangaName)
	req.URL.RawQuery = q.Encode()

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro ao buscar mangá: %s", resp.Status)
	}

	var result mangaSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("mangá não encontrado: %s", mangaName)
	}

	return result.Data[0].ID, nil
}

// getChapterID busca o ID do capítulo pelo número e ID do mangá
func (d *MangaDexDriver) getChapterID(mangaID, chapterNum string) (string, error) {
	req, err := http.NewRequest("GET", apiBaseURL+chapterEndpoint, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("manga", mangaID)
	q.Add("chapter", chapterNum)
	q.Add("translatedLanguage[]", d.language) // Filtrar por idioma português
	req.URL.RawQuery = q.Encode()

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro ao buscar capítulo: %s", resp.Status)
	}

	var result chapterResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("capítulo %s não encontrado em português", chapterNum)
	}

	return result.Data[0].ID, nil
}

// getChapterPages obtém as URLs das imagens de um capítulo
func (d *MangaDexDriver) getChapterPages(chapterID string) (*chapterPagesResult, error) {
	resp, err := d.client.Get(fmt.Sprintf("%s/at-home/server/%s", apiBaseURL, chapterID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao obter páginas: %s", resp.Status)
	}

	var result chapterPagesResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DownloadChapter baixa todas as páginas de um capítulo
func (d *MangaDexDriver) DownloadChapter(manga, chapter, outDir string) error {
	fmt.Printf("Buscando capítulo %s de %s em português (pt-br)...\n", chapter, manga)

	// 1. Encontrar o ID do mangá
	mangaID, err := d.findMangaID(manga)
	if err != nil {
		return err
	}

	// 2. Encontrar o ID do capítulo
	chapterID, err := d.getChapterID(mangaID, chapter)
	if err != nil {
		return err
	}

	// 3. Obter URLs das páginas
	pages, err := d.getChapterPages(chapterID)
	if err != nil {
		return err
	}

	fmt.Printf("Encontradas %d páginas. Iniciando download...\n", len(pages.Chapter.Data))

	// 4. Baixar cada página
	for i, filename := range pages.Chapter.Data {
		// Construir URL da imagem
		imageURL := fmt.Sprintf("%s/data/%s/%s", pages.BaseURL, pages.Chapter.Hash, filename)

		// Definir caminho de saída
		outPath := filepath.Join(outDir, fmt.Sprintf("%03d_%s", i+1, filename))

		fmt.Printf("Baixando página %d/%d...\n", i+1, len(pages.Chapter.Data))

		// Baixar a imagem
		if err := d.downloadImage(imageURL, outPath); err != nil {
			return fmt.Errorf("erro ao baixar página %d: %w", i+1, err)
		}

		// Esperar um pouco para não sobrecarregar o servidor
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("Download do capítulo %s concluído com sucesso!\n", chapter)
	return nil
}

// SetLanguage define o idioma das traduções
func (d *MangaDexDriver) SetLanguage(lang string) {
	d.language = lang
}

// downloadImage baixa uma imagem de uma URL e salva em um arquivo
func (d *MangaDexDriver) downloadImage(imageURL, outPath string) error {
	resp, err := d.client.Get(imageURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro ao baixar imagem: %s", resp.Status)
	}

	// Criar arquivo de saída
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copiar dados da resposta para o arquivo
	_, err = io.Copy(outFile, resp.Body)
	return err
}
