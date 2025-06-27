package crawler

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

// Crawler represents a web crawler instance
type Crawler struct {
	client *resty.Client
}

// New creates a new crawler instance
func New() *Crawler {
	return &Crawler{
		client: resty.New(),
	}
}

// Crawl fetches and parses a webpage
func (c *Crawler) Crawl(url string) error {
	resp, err := c.client.R().Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Basic example: extract title
	title := doc.Find("title").Text()
	fmt.Printf("Title: %s\n", title)

	return nil
}
