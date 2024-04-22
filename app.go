package main

import (
	"batch-downloader/core"
	"batch-downloader/core/downloader"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var Mutex = &sync.RWMutex{}
var newItems = map[string]downloader.DownloadItem{}

func main() {
    urls := GetUrlsFromFile("./files.txt")
    numUrls := len(urls)
    fmt.Println("Found", numUrls, "urls")
    bufferSize := runtime.NumCPU() * numUrls
    messages := make(chan downloader.DownloadItem, bufferSize)

    var wg sync.WaitGroup
    for _, Url := range urls {
        parsed, err := url.Parse(Url)
        core.Check(err)
        pathSegments := strings.Split(parsed.Path, "/")
        fileName := pathSegments[len(pathSegments)-1]
        fmt.Println(fileName)
        wg.Add(1)
        go downloader.HandleDownload(Url, fileName, &wg, messages)
    }
    for {
        time.Sleep(20 * time.Millisecond)
        fmt.Printf("\n----------------------\n")
        os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")
        _, err := downloader.GetDownloads()
        if(err != nil){
            fmt.Println("All downloads finished...")
            break
        }
        totalMessages := len(messages)
        for len(messages) > 0 {
            msg := <- messages
            newItems[msg.Url] = msg
        }
        sortedKeys := GetSortedListOfKeysFromMap(newItems)
        fmt.Printf("Downloading %v items... (Processed %v (bufferSize=%v) messages since last update)\n", len(sortedKeys), totalMessages, bufferSize)
        fmt.Println("-----------------")
        for _,key := range sortedKeys {
            item := newItems[key]
            fmt.Printf("%v    :   %.3fMB OF %.3fMB\n", item.FileName, AsMegabytes(item.BytesDownloaded), AsMegabytes(item.TotalSize))
        }
    }

    wg.Wait()
    close(messages)
    fmt.Println("Finished")
}

func AsMegabytes(bytes int64) float64 {
    return float64(float64(bytes) / (1024 * 1024))
}

func GetUrlsFromFile(path string) []string {
    dat, err := os.ReadFile(path)
    core.Check(err)
    asText := string(dat)
    lines := strings.Fields(asText)
    return lines
}

func GetSortedListOfKeysFromMap(mapToSort map[string]downloader.DownloadItem) []string {
    keys := make([]string, 0, len(mapToSort))

    for k := range mapToSort{
        keys = append(keys, k)
    }
    sort.Strings(keys)

    return keys    
}