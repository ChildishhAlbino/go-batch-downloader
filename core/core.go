package core

import (
	"fmt"
	"os"
	"runtime"
	"time"
)
var ErrorChannel = make(chan string, runtime.NumCPU())
var errorLogFilePath = "./error-log.log"
func Check(err error){
    if(err != nil){
        ErrorChannel <- err.Error()
        time.Sleep(2 * time.Second)
        panic(err)
    }
}

func HandleErrorLog(){
	f, _ := os.OpenFile(errorLogFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
    for msg := range ErrorChannel {
        fmt.Println(msg)
        f.WriteString(fmt.Sprintf("%s:: %s ", time.Now().Format(time.RFC3339), msg))
    }
}