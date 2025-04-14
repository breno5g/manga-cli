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
func (d *UnionMangasDriver) GetChapters(manga string) ([]string, error) {
	// Formatar o nome do mangá para URL (substituir espaços por hífen e converter para minúsculas)
	mangaURL := d.formatMangaURL(manga)

	// Requisição para a página do mangá
	resp, err := d.client.Get(fmt.Sprintf("%s/manga/%s", d.baseURL, mangaURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao acessar a página do mangá: %s", resp.Status)
	}

	// Parsear o HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extrair capítulos
	chapters := []string{}
	doc.Find(".chapters .cap").Each(func(i int, s *goquery.Selection) {
		// Extrair o número do capítulo do texto
		chapterText := strings.TrimSpace(s.Text())
		re := regexp.MustCompile(`Capítulo (\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(chapterText)
		if len(matches) > 1 {
			chapters = append(chapters, matches[1])
		}
	})

	return chapters, nil
}

// formatMangaURL formata o nome do mangá para uso em URLs
func (d *UnionMangasDriver) formatMangaURL(manga string) string {
	// Converter para minúsculas e substituir espaços por hífens
	formatted := strings.ToLower(manga)
	formatted = strings.ReplaceAll(formatted, " ", "-")
	// Remover caracteres especiais
	re := regexp.MustCompile(`[^a-z0-9\-]`)
	formatted = re.ReplaceAllString(formatted, "")
	return formatted
}

// DownloadChapter baixa um capítulo específico
func (d *UnionMangasDriver) DownloadChapter(manga, chapter, outDir string) error {
	// Formatar o nome do mangá para URL
	mangaURL := d.formatMangaURL(manga)

	// URL do capítulo
	chapterURL := fmt.Sprintf("%s/leitor/%s/%s", d.baseURL, mangaURL, chapter)

	// Acessar a página do capítulo
	resp, err := d.client.Get(chapterURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro ao acessar capítulo: %s", resp.Status)
	}

	// Parsear o HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// Extrair URLs das imagens
	var imageURLs []string
	doc.Find(".img-manga").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			imageURLs = append(imageURLs, src)
		}
	})

	if len(imageURLs) == 0 {
		return fmt.Errorf("nenhuma imagem encontrada no capítulo")
	}

	// Baixar cada imagem
	for i, imageURL := range imageURLs {
		// Definir caminho de saída
		ext := filepath.Ext(imageURL)
		if ext == "" {
			ext = ".jpg" // Extensão padrão se não for detectada
		}
		outPath := filepath.Join(outDir, fmt.Sprintf("%03d%s", i+1, ext))

		// Baixar a imagem
		if err := d.downloadImage(imageURL, outPath); err != nil {
			return fmt.Errorf("erro ao baixar página %d: %w", i+1, err)
		}

		// Esperar um pouco para não sobrecarregar o servidor
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// downloadImage baixa uma imagem e salva em um arquivo
func (d *UnionMangasDriver) downloadImage(imageURL, outPath string) error {
	// Configurar cabeçalhos para evitar bloqueio
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return err
	}

	// Adicionar cabeçalhos comuns para evitar bloqueios
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", d.baseURL)

	// Fazer a requisição
	resp, err := d.client.Do(req)
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
