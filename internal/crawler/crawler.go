package crawler

import (
	"context"
	"errors"
	"io"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/time/rate"
)

var errNotHttp = errors.New("non-http scheme")

type Storer interface {
	StorerPage(ctx context.Context, page *Page) (int64, error)
}

type Fetcher interface {
	Fatch(ctx context.Context, url string) (context io.ReadCloser, err error)
}

type Page struct {
	ID    int64
	URL   string
	Title string
	Body  string //без HTML тегов
}

type Crawler struct {
	jobs    chan string
	results chan *Page
	wg      sync.WaitGroup
	limiter *rate.Limiter
	storage Storer
	fetcher Fetcher
	visited *VisitedCache

	workers int
}

func NewCrawler(workers int, requestsPerSec int, storage Storer, fetcher Fetcher) *Crawler {
	limiter := rate.NewLimiter(rate.Every(time.Second/time.Duration(requestsPerSec)), 1)

	return &Crawler{
		jobs:    make(chan string, workers*2),
		results: make(chan *Page, workers*2),
		limiter: limiter,
		storage: storage,
		fetcher: fetcher,
		workers: workers,
		visited: NewVisitedCache(),
	}
}

func (c *Crawler) Start(ctx context.Context, seedURLs []string) {
	c.wg.Add(1)
	go c.processResults(ctx)

	for i := 1; i <= c.workers; i++ {
		c.wg.Add(1)
		go c.worker(ctx, i)
	}

	for _, u := range seedURLs {
		c.AddJob(u)
	}
}

func (c *Crawler) Stop() {
	close(c.jobs)
	c.wg.Wait()
	close(c.results)
}

func (c *Crawler) AddJob(url string) {
	c.jobs <- url
}

func (c *Crawler) worker(ctx context.Context, id int) {
	defer c.wg.Done()
	log.Printf("Воркер %d запущен", id)

	for jobURL := range c.jobs {
		if !c.visited.AddIfNotExists(jobURL) {
			continue
		}

		log.Printf("Воркер %d: обрабатывает %s", id, jobURL)

		if err := c.limiter.Wait(ctx); err != nil {
			log.Printf("Воркер %d остановлен из-за ошибки ограничителя: %v", id, err)
			break
		}

		body, err := c.fetcher.Fetch(ctx, jobURL)
		if err != nil {
			log.Printf("Ошибка загрузки URL %s: %v", jobURL, err)
			continue
		}

		title, text, links := c.parseHTML(jobURL, body)
		body.Close()

		c.results <- &Page{
			URL:   jobURL,
			Title: title,
			Body:  text,
		}

		for _, link := range links {
			c.AddJob(link)
		}
	}
	log.Printf("Воркер %d завершает работу", id)
}

func (c *Crawler) processResults(ctx context.Context) {
	defer c.wg.Done()

	for page := range c.results {
		if _, err := c.storage.StorePage(ctx, page); err != nil {
			log.Printf("Ошибка сохранения страницы %s: %v", page.URL, err)
		} else {
			log.Printf("Страница %s успешно сохранена", page.URL)
		}
	}
}

func (c *Crawler) parseHTML(baseURL string, body io.Reader) (string, string, []string) {
	doc, err := html.Parse(body)
	if err != nil {
		log.Printf("Ошибка парсинга HTML для %s: %v", baseURL, err)
		return "", "", nil
	}

	var title string
	var text strings.Builder
	var links []string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "title" && n.FirstChild != nil {
				title = n.FirstChild.Data
			}
			if n.Data == "script" || n.Data == "style" {
				return
			}
			if n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "href" {
						if resolvedURL, err := resolveURL(baseURL, a.Val); err == nil {
							links = append(links, resolvedURL)
						}
						break
					}
				}
			}
		} else if n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteString(" ")
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			f(child)
		}
	}

	f(doc)
	return strings.TrimSpace(title), strings.Join(strings.Fields(text.String()), " "), links
}

func resolveURL(base, relative string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	relativeURL, err := url.Parse(relative)
	if err != nil {
		return "", err
	}

	if relativeURL.Scheme != "" && relativeURL.Scheme != "http" && relativeURL.Scheme != "https" {
		return "", errNotHttp
	}

	relativeURL.Fragment = ""

	return baseURL.ResolveReference(relativeURL).String(), nil
}
