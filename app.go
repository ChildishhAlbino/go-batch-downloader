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

var itemsStillDownloading = map[string]downloader.DownloadItem{}
var longestFileName int = 0
var outputChannel = make(chan string, runtime.NumCPU())
var lastOutputTimestamp = time.Now()

func main() {
	var wg sync.WaitGroup
	start := time.Now()
	go core.HandleErrorLog()

	urls := GetUrlsFromFile("./files.txt")
	numUrls := len(urls)
	bufferSize := runtime.NumCPU() * numUrls * 10
	messageChannel := make(chan downloader.DownloadItem, bufferSize)
	
	for _, Url := range urls {
		parsed, err := url.Parse(Url)
		core.Check(err)
		pathSegments := strings.Split(parsed.Path, "/")
		fileName := pathSegments[len(pathSegments)-1]
		length := len(fileName)
		if length > longestFileName {
			longestFileName = length
		}
		wg.Add(1)
		go downloader.HandleDownload(Url, fileName, &wg, messageChannel)
		outputChannel <- GetOutputText(start)
	}
    go func () {
        for msg := range outputChannel {
            os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")      
            fmt.Print(msg)
        }
    }()
	
	go func() {
		wg.Wait()
		close(messageChannel)
        close(outputChannel)
		close(core.ErrorChannel)
	}()

	for msg := range messageChannel {
		_, err := downloader.GetDownloads()
		if err != nil {
			fmt.Println("All downloads finished...")
			break
		}
		if msg.BytesDownloaded == msg.TotalSize {
			delete(itemsStillDownloading, msg.Url)
			continue
		}
		itemsStillDownloading[msg.Url] = msg
        
        secondsSinceLastUpdate := time.Since(lastOutputTimestamp).Milliseconds()
        if(secondsSinceLastUpdate > 200) {
            outputChannel <- GetOutputText(start)
            lastOutputTimestamp = time.Now()
        }
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

func GetOutputText(start time.Time) string {
    output := ""
    sortedKeys := GetSortedListOfKeysFromMap(itemsStillDownloading)
    output += fmt.Sprintf("Downloading %v items...\n", len(sortedKeys))
    output += fmt.Sprintln(strings.Repeat("-", longestFileName+2))
    for _, key := range sortedKeys {
        item := itemsStillDownloading[key]
        output += fmt.Sprintf("%-*s: %.2fMB / %.2fMB\n", longestFileName+2, item.FileName, AsMegabytes(item.BytesDownloaded), AsMegabytes(item.TotalSize))
    }
    output += fmt.Sprintln(strings.Repeat("-", longestFileName+2))
    elapsed := time.Since(start).Round(time.Millisecond)
    output += fmt.Sprintf("Total duration: %s\n", elapsed)
    return output
}
