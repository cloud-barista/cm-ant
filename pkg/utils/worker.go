package utils

import (
	"log"
	"time"
)

type TempFileRemoveWorker struct {
	Stopped         bool
	ShutdownChannel chan string
	Interval        time.Duration
}

func NewWorker(interval time.Duration) *TempFileRemoveWorker {
	return &TempFileRemoveWorker{
		Stopped:         false,
		ShutdownChannel: make(chan string),
		Interval:        interval,
	}
}

func (w *TempFileRemoveWorker) Run() {

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

func (w *TempFileRemoveWorker) Shutdown() {
	w.Stopped = true

	w.ShutdownChannel <- "Down"
	<-w.ShutdownChannel

	close(w.ShutdownChannel)
}

// temp folder 하위 폴더명을 확인하여 현재의 unix 시간과 시간을 확인해서 .5 시간 이상 차이나는 데이터는 삭제
func (w *TempFileRemoveWorker) Action() {
	time.Sleep(5 * time.Second)
	log.Println("Action complete!")
}
