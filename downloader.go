package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

const MAX_NAME_CHARACTERS int = 17

func NewDownloader() *Downloader {
	return &Downloader{
		Channel:  make(chan int, 2),
		Progress: mpb.New(mpb.WithWidth(60), mpb.WithRefreshRate(180*time.Millisecond)),
		client:   &http.Client{},
	}
}

type Downloader struct {
	Progress                   *mpb.Progress
	Channel                    chan int
	wg                         sync.WaitGroup
	OnDownloadCompleteCallback func(string)
	client                     *http.Client
}

func (d *Downloader) GetProgressBar(filename string, size int64) *mpb.Bar {
	var bar *mpb.Bar
	if len(filename) > MAX_NAME_CHARACTERS {
		marquee := NewChangeNameDecor(filename, MAX_NAME_CHARACTERS)
		bar = d.Progress.AddBar(size, mpb.BarStyle("[=>-|"),
			mpb.PrependDecorators(
				decor.Name("[ "),
				marquee.MarqueeText(decor.WC{W: 5, C: decor.DidentRight}),
				decor.Name(" ] "),
				decor.CountersKibiByte("% .2f / % .2f"),
			),
			mpb.AppendDecorators(
				decor.AverageETA(decor.ET_STYLE_GO),
				decor.Name("]"),
				decor.AverageSpeed(decor.UnitKiB, " % .2f"),
			),
		)
	} else {
		bar = d.Progress.AddBar(size, mpb.BarStyle("[=>-|"),
			mpb.PrependDecorators(
				decor.Name("[ "),
				decor.Name(filename, decor.WC{W: 5, C: decor.DidentRight}),
				decor.Name(" ] "),
				decor.CountersKibiByte("% .2f / % .2f"),
			),
			mpb.AppendDecorators(
				decor.AverageETA(decor.ET_STYLE_GO),
				decor.Name("]"),
				decor.AverageSpeed(decor.UnitKiB, " % .2f"),
			),
		)
	}
	return bar
}

func (d *Downloader) SetConcurrency(con int) {
	d.Channel = make(chan int, con)
}

func (d *Downloader) SetOnDownloadCompleteCallback(callback func(string)) {
	d.OnDownloadCompleteCallback = callback
}

func (d *Downloader) AddDownload(link string, filename string) {
	res, err := d.client.Get(link)
	if err != nil {
		fmt.Println(err)
		return
	}
	d.Channel <- 1
	d.wg.Add(1)
	go func(res *http.Response) {
		defer func() {
			<-d.Channel
			d.wg.Done()
		}()
		writer, err := os.OpenFile(Uniquify(filename), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer writer.Close()
		size := GetSizeByHeader(res.Header)
		bar := d.GetProgressBar(GetFileNameUrl(filename), size)

		proxyReader := bar.ProxyReader(res.Body)
		io.Copy(writer, proxyReader)
		proxyReader.Close()
		bar.Abort(true)
		if d.OnDownloadCompleteCallback != nil {
			d.OnDownloadCompleteCallback(filename)
		}
	}(res)
}

func (d *Downloader) WaitDownload() {
	d.wg.Wait()
}
