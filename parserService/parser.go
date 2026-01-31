package parserService

import (
	"context"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ParserService interface {
	Parse(ctx context.Context, url string) (*ParsedPage, error)
}
type ParsedPage struct {
	URL         string
	Title       string
	Description string
	ImageURL    string
	SiteName    string
}

type parserService struct{}

func New() ParserService {
	return &parserService{}
}

func (p *parserService) Parse(ctx context.Context, url string) (*ParsedPage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Telegram Bot)")

	client := &http.Client{Timeout: time.Second * 5}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return nil, err
	}

	page := &ParsedPage{URL: url}

	page.Title = doc.Find("meta[property='og:title']").AttrOr("content", "")

	if page.Title == "" {
		page.Title = doc.Find("title").Text()
	}

	page.Description = doc.Find("meta[property='og:description']").AttrOr("content", "")

	page.SiteName = doc.Find("meta[property='og:image']").AttrOr("content", "")

	return page, nil
}
