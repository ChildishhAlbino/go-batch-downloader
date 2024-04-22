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
var longestFileName int = 0

func main() {
	start := time.Now()
	urls := GetUrlsFromFile("./files.txt")
	numUrls := len(urls)
	fmt.Println("Found", numUrls, "urls")
	bufferSize := runtime.NumCPU() * numUrls
	messageChannel := make(chan downloader.DownloadItem, bufferSize)

	var wg sync.WaitGroup
	for _, Url := range urls {
		parsed, err := url.Parse(Url)
		core.Check(err)
		pathSegments := strings.Split(parsed.Path, "/")
		fileName := pathSegments[len(pathSegments)-1]
		length := len(fileName)
		if length > longestFileName {
			longestFileName = length
		}
		fmt.Println(fileName)
		wg.Add(1)
		go downloader.HandleDownload(Url, fileName, &wg, messageChannel)
	}

	go func() {
		wg.Wait()
		close(messageChannel)
	}()

	for msg := range messageChannel {
		os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")
		_, err := downloader.GetDownloads()
		if err != nil {
			fmt.Println("All downloads finished...")
			break
		}
		if msg.BytesDownloaded == msg.TotalSize {
			delete(newItems, msg.Url)
			continue
		}
		newItems[msg.Url] = msg
		sortedKeys := GetSortedListOfKeysFromMap(newItems)
		fmt.Printf("Downloading %v items...\n", len(sortedKeys))
		fmt.Println(strings.Repeat("-", longestFileName+2))
		for _, key := range sortedKeys {
			item := newItems[key]
			fmt.Printf("%-*s: %.2fMB / %.2fMB\n", longestFileName+2, item.FileName, AsMegabytes(item.BytesDownloaded), AsMegabytes(item.TotalSize))
		}
		fmt.Println(strings.Repeat("-", longestFileName+2))
		elapsed := time.Since(start).Round(time.Second)
		fmt.Printf("Total duration: %s\n", elapsed)
	}

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

	for k := range mapToSort {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
