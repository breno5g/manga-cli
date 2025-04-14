// drivers/mangadex/driver.go
package mangadex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
		client:   &http.Client{},
		language: "pt-br", // Define português do Brasil como idioma padrão
	}
}

// API de busca do MangaDex (simplificada)
const (
	apiBaseURL      = "https://api.mangadex.org"
	mangaSearchPath = "/manga"
	chapterPath     = "/chapter"
	atHomeServer    = "/at-home/server"
)

// mangaSearchResult representa a resposta da API de busca do MangaDex
type mangaSearchResult struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// chapterResult representa a resposta da API de capítulos do MangaDex
type chapterResult struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Chapter string `json:"chapter"`
		} `json:"attributes"`
	} `json:"data"`
}

// chapterPagesResult representa a resposta com os dados de páginas de um capítulo
type chapterPagesResult struct {
	BaseURL string   `json:"baseUrl"`
	Chapter struct {
		Hash      string   `json:"hash"`
		Data      []string `json:"data"`
		DataSaver []string `json:"dataSaver"`
	} `json:"chapter"`
}

// GetChapters retorna a lista de capítulos disponíveis para o mangá
func (d *MangaDexDriver) GetChapters(mangaName string) ([]string, error) {
	mangaID, err := d.findMangaID(mangaName)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar ID do mangá: %w", err)
	}

	u, err := url.Parse(apiBaseURL + chapterPath)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Add("manga", mangaID)
	q.Add("translatedLanguage[]", d.language)
	q.Add("limit", "100")
	u.RawQuery = q.Encode()

	resp, err := d.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result chapterResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	chapters := make([]string, 0, len(result.Data))
	for _, chapter := range result.Data {
		if chapter.Attributes.Chapter != "" {
			chapters = append(chapters, chapter.Attributes.Chapter)
		}
	}

	return chapters, nil
}

// findMangaID busca o ID do mangá pelo nome
func (d *MangaDexDriver) findMangaID(mangaName string) (string, error) {
	u, err := url.Parse(apiBaseURL + mangaSearchPath)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("title", mangaName)
	u.RawQuery = q.Encode()

	resp, err := d.client.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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
func (d *MangaDexDriver) getChapterID(mangaID, chapterNumber string) (string, error) {
	u, err := url.Parse(apiBaseURL + chapterPath)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("manga", mangaID)
	q.Add("chapter", chapterNumber)
	q.Add("translatedLanguage[]", d.language)
	u.RawQuery = q.Encode()

	resp, err := d.client.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result chapterResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("capítulo não encontrado: %s", chapterNumber)
	}

	return result.Data[0].ID, nil
}

// getChapterPages obtém as URLs das imagens de um capítulo
func (d *MangaDexDriver) getChapterPages(chapterID string) (*chapterPagesResult, error) {
	resp, err := d.client.Get(fmt.Sprintf("%s%s/%s", apiBaseURL, atHomeServer, chapterID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result chapterPagesResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DownloadChapter baixa todas as páginas de um capítulo
func (d *MangaDexDriver) DownloadChapter(mangaName, chapterNumber, outputDir string) error {
	mangaID, err := d.findMangaID(mangaName)
	if err != nil {
		return err
	}

	chapterID, err := d.getChapterID(mangaID, chapterNumber)
	if err != nil {
		return err
	}

	pages, err := d.getChapterPages(chapterID)
	if err != nil {
		return err
	}

	for i, page := range pages.Chapter.Data {
		imageURL := fmt.Sprintf("%s/data/%s/%s", pages.BaseURL, pages.Chapter.Hash, page)
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%03d%s", i+1, filepath.Ext(page)))

		if err := d.downloadImage(imageURL, outputPath); err != nil {
			return fmt.Errorf("erro ao baixar página %d: %w", i+1, err)
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// SetLanguage define o idioma das traduções
func (d *MangaDexDriver) SetLanguage(language string) {
	d.language = language
}

// downloadImage baixa uma imagem de uma URL e salva em um arquivo
func (d *MangaDexDriver) downloadImage(imageURL, outputPath string) error {
	resp, err := d.client.Get(imageURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro ao baixar imagem: status %d", resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
