package downloader

import (
	"batch-downloader/core"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

var downloads map[string]DownloadItem = make(map[string]DownloadItem)
var downloadMapMutex = &sync.RWMutex{}

type ProgressReader struct {
	Reader io.Reader
	Size   int64
	Pos    int64
	Url    string
	Ch     chan<- DownloadItem
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	shouldProcess := err == nil || err.Error() == "EOF"
	if shouldProcess {
		pr.Pos += int64(n)
		existing := downloads[pr.Url]
		existing.BytesDownloaded = pr.Pos
		existing.TotalSize = pr.Size
		pr.Ch <- existing
	} else {
		fmt.Println(err)
	}
	return n, err
}

func GetDownloads() (map[string]DownloadItem, error) {
	if len(downloads) == 0 {
		return nil, errors.New("all downloads are closed")
	}
	return downloads, nil
}

func HandleDownload(url string, outName string, wg *sync.WaitGroup, ch chan<- DownloadItem) error {
	downloadMapMutex.Lock()
	downloads[url] = DownloadItem{FileName: outName, Url: url}
	downloadMapMutex.Unlock()
	defer wg.Done()
	tempPath := fmt.Sprintf(".tmp_%s", outName)
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		log.Fatalf("Error while downloading: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	outputFile, _ := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, 0644)

	progressReader := &ProgressReader{
		Reader: resp.Body,
		Size:   resp.ContentLength,
		Url:    url,
		Ch:     ch,
	}

	if _, err := io.Copy(outputFile, progressReader); err != nil {
		log.Fatalf("Error while downloading: %v", err)
		core.ErrorChannel <- err.Error()
		return err
	}
	// Do this instead of deferring closure as Windows doesn't allow renames while another process has the file 'open'
	outputFile.Close()
	err := os.Rename(tempPath, outName)
	core.Check(err)
	downloadMapMutex.Lock()
	delete(downloads, url)
	downloadMapMutex.Unlock()
	return nil
}

type DownloadItem struct {
	FileName        string
	Url             string
	BytesDownloaded int64
	TotalSize       int64
}
