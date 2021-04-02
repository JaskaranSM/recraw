package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func SanitizeURL(url string) string {
	if strings.HasSuffix(url, "/") {
		return string(url[:len(url)-1])
	}
	return url
}

func GetReaderByLink(link string) (io.Reader, error) {
	res, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func GetMimeByReader(rdr io.Reader) (*mimetype.MIME, error) {
	return mimetype.DetectReader(rdr)
}

func GetSizeByHeader(header http.Header) int64 {
	size, _ := strconv.Atoi(header.Get("Content-Length"))
	return int64(size)
}

func GetFileNameUrl(link string) string {
	strip := func(l string) string {
		data := strings.Split(l, "/")
		return strings.ReplaceAll(data[len(data)-1], "+", " ")
	}
	parsed, err := url.Parse(link)
	if err != nil {
		return strip(link)
	}
	return strip(parsed.EscapedPath())
}

func ParseCommandString(filename, cmd string) string {
	return strings.ReplaceAll(cmd, "{filename}", fmt.Sprintf("'%s'", filename))
}

func Uniquify(filename string) string {
	index := 1
	extension := filepath.Ext(filename)
	name := filename[0 : len(filename)-len(extension)]
	for {
		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			return filename
		}
		filename = fmt.Sprintf("%s (%d)%s", name, index, extension)
		index += 1
	}
	return filename
}

func GetUrlHostName(link string) string {
	parsed, err := url.Parse(link)
	if err != nil {
		return link
	}
	return parsed.Scheme + "://" + parsed.Hostname()
}
