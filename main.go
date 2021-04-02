// recraw project main.go
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"
)

func StringToStringSlice(str string) []string {
	return strings.Split(str, " ")
}

func callback(c *cli.Context) error {
	ignoredExts := c.String("ignore")
	dlExts := c.String("dl")
	link := c.String("link")
	if link == "" {
		return errors.New("try --help")
	}
	downloadCompleteCmd := c.String("on-download-complete")
	opts := NewScraperOptions()
	opts.Conncurrency = c.Int("con")
	opts.DownloadDir = c.String("dldir")
	for _, i := range StringToStringSlice(ignoredExts) {
		opts.AddIgnoreExtension(i)
	}
	for _, i := range StringToStringSlice(dlExts) {
		opts.AddDownloadExtension(i)
	}
	scraper := NewScraper(opts)
	if downloadCompleteCmd != "" {
		scraper.SetOnDownloadCompleteCallback(downloadCompleteCmd)
	}
	scraper.Dlr.SetConcurrency(opts.Conncurrency)
	scraper.Scrape(link)
	scraper.Wait()
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "ReCraw - recursive crawler"
	app.Usage = "A recursive web crawler written in Go."
	app.UsageText = fmt.Sprintf("%s [global options] [arguments...]", os.Args[0])
	app.Authors = []cli.Author{
		{Name: "JaskaranSM"},
	}
	app.Action = callback
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "ignore",
			Value: "",
			Usage: "specify extensions to ignore by space (. is mandatory)",
		},
		&cli.StringFlag{
			Name:  "dl",
			Value: "",
			Usage: "specify extensions to download by space (. is mandatory)",
		},
		&cli.StringFlag{
			Name:  "link",
			Value: "",
			Usage: "link to site",
		},
		&cli.StringFlag{
			Name:  "dldir",
			Value: ".",
			Usage: "download directory",
		},
		&cli.IntFlag{
			Name:  "con",
			Value: 1,
			Usage: "concurrency",
		},
		&cli.StringFlag{
			Name:  "on-download-complete",
			Value: "",
			Usage: "specify command to execute after download complete of each file, {filename} will be parsed and fed into command, Note:- doesnt work on windows",
		},
	}
	app.Version = "0.1"
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
