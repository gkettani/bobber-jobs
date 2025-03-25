package fetcher

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// BaseFetcher provides common HTTP functionality
// it encapsulates all the logic to fetch content from a URL
type BaseFetcher struct {
	httpClient *http.Client
}

func NewBaseFetcher() *BaseFetcher {
	return &BaseFetcher{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchContent fetches raw content from a URL
func (f *BaseFetcher) fetchContent(url string) ([]byte, error) {
	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (f *BaseFetcher) fetchHTML(url string) (*goquery.Document, error) {
	content, err := f.fetchContent(url)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(content))
}
