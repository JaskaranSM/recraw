// scraper.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
)

type ScraperOptions struct {
	IgnoreExtensions   []string
	DownloadExtensions []string
	DownloadDir        string
	Conncurrency       int
}

func (s *ScraperOptions) AddIgnoreExtension(ext string) {
	s.IgnoreExtensions = append(s.IgnoreExtensions, ext)
}

func (s *ScraperOptions) AddDownloadExtension(ext string) {
	s.DownloadExtensions = append(s.DownloadExtensions, ext)
}

func (s *ScraperOptions) IsExtIgnored(ext string) bool {
	for _, i := range s.IgnoreExtensions {
		if ext == i {
			return true
		}
	}
	return false
}

func (s *ScraperOptions) ShouldDownload(ext string) bool {
	for _, i := range s.DownloadExtensions {
		if ext == i {
			return true
		}
	}
	return false
}

func NewScraperOptions() *ScraperOptions {
	return &ScraperOptions{DownloadDir: "."}
}

type Scraper struct {
	Opts           *ScraperOptions
	Dlr            *Downloader
	DownloadedUrls []string
}

func (s *Scraper) AddDownloadUrl(url string) {
	s.DownloadedUrls = append(s.DownloadedUrls, url)
}

func (s *Scraper) IsUrlDownloaded(url string) bool {
	for _, i := range s.DownloadedUrls {
		if i == url {
			return true
		}
	}
	return false
}

func (s *Scraper) GetCollector() *colly.Collector {
	c := colly.NewCollector()
	c.AllowURLRevisit = false
	c.MaxDepth = 5
	c.OnError(func(_ *colly.Response, err error) {

	})
	return c
}

func (s *Scraper) GetQueue(con int) *queue.Queue {
	q, _ := queue.New(
		con,
		&queue.InMemoryQueueStorage{MaxSize: 100000},
	)
	return q
}

func (s *Scraper) Scrape(base string) {
	os.MkdirAll(s.Opts.DownloadDir, 0755)
	base = SanitizeURL(base)
	c := s.GetCollector()
	q := s.GetQueue(s.Opts.Conncurrency)
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasPrefix(link, "/") || !strings.HasPrefix(link, "http") {
			link = e.Request.URL.String() + link
		}
		fmt.Println(link)
		do := false
		for _, proto := range []string{"http", "https", "ftp"} {
			if strings.HasPrefix(link, proto) {
				do = true
			}
		}
		if !do {
			return
		}
		rdr, err := GetReaderByLink(link)
		if err != nil {
			fmt.Println(err)
			return
		}
		mime, err := GetMimeByReader(rdr)
		if err != nil {
			fmt.Println(err)
			return
		}
		if s.Opts.IsExtIgnored(mime.Extension()) {
			return
		}
		if s.Opts.ShouldDownload(mime.Extension()) && !s.IsUrlDownloaded(link) {
			s.AddDownloadUrl(link)
			go s.Dlr.AddDownload(link, path.Join(s.Opts.DownloadDir, GetFileNameUrl(link)))
			return
		}
		req, err := e.Request.New("GET", link, nil)
		if err != nil {
			fmt.Println(err)
		} else {
			q.AddRequest(req)
		}
	})
	q.AddURL(base)
	q.Run(c)
}

func (s *Scraper) SetOnDownloadCompleteCallback(cmd string) {
	s.Dlr.SetOnDownloadCompleteCallback(func(filename string) {
		exec.Command("bash", "-c", ParseCommandString(filename, cmd)).Run()
	})
}

func (s *Scraper) Wait() {
	s.Dlr.WaitDownload()
}

func NewScraper(opts *ScraperOptions) *Scraper {
	return &Scraper{Opts: opts, Dlr: NewDownloader()}
}
