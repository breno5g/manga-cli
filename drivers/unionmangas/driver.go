// drivers/unionmangas/driver.go
package unionmangas

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/breno5g/manga-cli/interfaces"
)

// UnionMangasDriver implementa a interface Driver para o site UnionMangas
type UnionMangasDriver struct {
	client  *http.Client
	baseURL string
}

// NewDriver cria uma nova instância do driver UnionMangas
func NewDriver() interfaces.Driver {
	return &UnionMangasDriver{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://unionleitor.top",
	}
}

// GetChapters retorna a lista de capítulos disponíveis para um mangá
func (d *UnionMangasDriver) GetChapters(mangaName string) ([]string, error) {
	mangaURL := d.formatMangaURL(mangaName)
	url := fmt.Sprintf("%s/manga/%s", d.baseURL, mangaURL)

	resp, err := d.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao acessar a página do mangá: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var chapters []string
	doc.Find(".capitulos a").Each(func(i int, s *goquery.Selection) {
		chapterText := s.Text()
		re := regexp.MustCompile(`Cap[íi]tulo\s+(\d+(?:\.\d+)?)`)
		if matches := re.FindStringSubmatch(chapterText); len(matches) > 1 {
			chapters = append(chapters, matches[1])
		}
	})

	return chapters, nil
}

// formatMangaURL formata o nome do mangá para uso em URLs
func (d *UnionMangasDriver) formatMangaURL(mangaName string) string {
	name := strings.ToLower(mangaName)
	name = strings.ReplaceAll(name, " ", "-")
	name = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(name, "")
	return name
}

// DownloadChapter baixa um capítulo específico
func (d *UnionMangasDriver) DownloadChapter(mangaName, chapterNumber, outputDir string) error {
	mangaURL := d.formatMangaURL(mangaName)
	url := fmt.Sprintf("%s/leitor/%s/%s", d.baseURL, mangaURL, chapterNumber)

	resp, err := d.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro ao acessar capítulo: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	var imageURLs []string
	doc.Find("#images img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			imageURLs = append(imageURLs, src)
		}
	})

	if len(imageURLs) == 0 {
		return fmt.Errorf("nenhuma imagem encontrada no capítulo")
	}

	for i, imageURL := range imageURLs {
		ext := filepath.Ext(imageURL)
		if ext == "" {
			ext = ".jpg"
		}
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%03d%s", i+1, ext))

		if err := d.downloadImage(imageURL, outputPath); err != nil {
			return fmt.Errorf("erro ao baixar página %d: %w", i+1, err)
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// downloadImage baixa uma imagem e salva em um arquivo
func (d *UnionMangasDriver) downloadImage(imageURL, outputPath string) error {
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", d.baseURL)
	req.Header.Set("Accept", "image/webp,*/*")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
