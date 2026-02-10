package policy

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Fetcher downloads and extracts text from URLs and files.
type Fetcher struct {
	client *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// FetchURL downloads a URL and extracts plain text.
func (f *Fetcher) FetchURL(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 50*1024*1024)) // 50MB cap
	if err != nil {
		return "", err
	}

	contentType := resp.Header.Get("Content-Type")
	return extractText(body, contentType, url), nil
}

// ReadFile reads a local file and extracts plain text.
func (f *Fetcher) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	contentType := ""
	if strings.HasSuffix(strings.ToLower(path), ".pdf") {
		contentType = "application/pdf"
	} else if strings.HasSuffix(strings.ToLower(path), ".html") || strings.HasSuffix(strings.ToLower(path), ".htm") {
		contentType = "text/html"
	}

	return extractText(data, contentType, path), nil
}

func extractText(data []byte, contentType, source string) string {
	text := string(data)

	// Strip HTML tags if HTML content
	if strings.Contains(contentType, "text/html") || strings.HasSuffix(strings.ToLower(source), ".html") {
		text = stripHTML(text)
	}

	return strings.TrimSpace(text)
}

var (
	scriptRe = regexp.MustCompile(`(?si)<script.*?</script>`)
	styleRe  = regexp.MustCompile(`(?si)<style.*?</style>`)
	tagRe    = regexp.MustCompile(`<[^>]+>`)
	spaceRe  = regexp.MustCompile(`\s+`)
)

func stripHTML(html string) string {
	text := scriptRe.ReplaceAllString(html, "")
	text = styleRe.ReplaceAllString(text, "")
	text = tagRe.ReplaceAllString(text, " ")
	text = spaceRe.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
