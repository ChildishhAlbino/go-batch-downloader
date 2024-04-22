package downloader

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

var downloads map[string]DownloadItem = make(map[string]DownloadItem)
var Mutex = &sync.RWMutex{}
type ProgressReader struct {
    Reader io.Reader
    Size   int64
    Pos    int64
	Url    string
    Ch     chan<-DownloadItem
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
    n, err := pr.Reader.Read(p)
    if err == nil {
        pr.Pos += int64(n)
		existing := downloads[pr.Url]
        existing.BytesDownloaded = pr.Pos
        existing.TotalSize = pr.Size
        pr.Ch <- existing
    }
    return n, err
}

func GetDownloads() (map[string]DownloadItem, error) {
	if(len(downloads) == 0){
		return nil, errors.New("all downloads are closed")
	}
	return downloads, nil
}

func HandleDownload(url string, outName string, wg *sync.WaitGroup, ch chan<-DownloadItem) (error) {
	Mutex.Lock()
	downloads[url] = DownloadItem{FileName: outName, Url: url}
	Mutex.Unlock()
	defer wg.Done()
    tempPath := fmt.Sprintf(".tmp_%s", outName)
	fmt.Println("Downloading to temporary path", tempPath)
    req, _ := http.NewRequest("GET", url, nil)
    resp, _ := http.DefaultClient.Do(req)
    if resp.StatusCode != 200 {
        log.Fatalf("Error while downloading: %v", resp.StatusCode)
    }
    defer resp.Body.Close()

    f, _ := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, 0644)
    defer f.Close()

    progressReader := &ProgressReader{
        Reader: resp.Body,
        Size:   resp.ContentLength,
		Url: 	url,
        Ch:     ch,
    }

    if _, err := io.Copy(f, progressReader); err != nil {
        log.Fatalf("Error while downloading: %v", err)
		return err
    }

    os.Rename(tempPath, outName)
	Mutex.Lock()
	delete(downloads, url)
	Mutex.Unlock()
	return nil
}

type DownloadItem struct {
    FileName string
	Url string
    BytesDownloaded  int64
	TotalSize int64
}