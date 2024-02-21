package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Worker interface {
	Run()
	Shutdown()
	Action()
}

type tempFileRemoveWorker struct {
	Stopped         bool
	ShutdownChannel chan string
	Interval        time.Duration
}

func NewWorker(interval time.Duration) Worker {
	return &tempFileRemoveWorker{
		Stopped:         false,
		ShutdownChannel: make(chan string),
		Interval:        interval,
	}
}

func (w *tempFileRemoveWorker) Run() {

	log.Println("TempFileRemoveWorker Started")

	for {
		select {
		case <-w.ShutdownChannel:
			log.Println("worker shut down..")
			w.ShutdownChannel <- "Down"
			return
		default:
			log.Println("worker actions called")
		}

		w.Action()
		time.Sleep(w.Interval)

	}
}

func (w *tempFileRemoveWorker) Shutdown() {
	w.Stopped = true

	w.ShutdownChannel <- "Down"
	<-w.ShutdownChannel

	close(w.ShutdownChannel)
}

// temp folder 하위 폴더명을 확인하여 현재의 unix 시간과 시간을 확인해서 .5 시간 이상 차이나는 데이터는 삭제
func (w *tempFileRemoveWorker) Action() {
	currentTime := time.Now()
	currentTimestamp := currentTime.UnixMilli()

	files, err := os.ReadDir("temp")
	if err != nil {
		log.Printf("error while reading temp directory for remove; %v\n", err)
	}

	log.Printf("%d folders read from temp directory", len(files))
	standardMilliSec := int64(10 * 60 * 1000)
	for _, file := range files {
		folderName := file.Name()
		time := getFirstPart(folderName)
		timestamp, _ := strconv.Atoi(time)

		if err == nil && file.IsDir() && currentTimestamp-int64(timestamp) > standardMilliSec {
			os.RemoveAll(fmt.Sprintf("temp/%s", folderName))
			log.Printf("%s folder deleted in temp directory.\n", folderName)
		}
	}
}

func getFirstPart(input string) string {
	parts := strings.Split(input, "-")
	return parts[0]
}
