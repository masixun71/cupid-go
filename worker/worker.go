package worker

import (
	"../jobQueue"
	"fmt"
)

type Worker struct {
	JobChannel jobQueue.JobChan
	quit       chan bool
}

func NewWorker() *Worker  {

	return &Worker{
		JobChannel:make(chan jobQueue.Job),
		quit: make(chan bool),
	}
}


func (w *Worker) Start(workerPool jobQueue.WorkerChan) {
	workerPool <- w.JobChannel
	go func(workerPool jobQueue.WorkerChan) {
		for {
			select {
			case job := <-w.JobChannel:
				if err := job.Do(); err != nil {
					fmt.Printf("excute job failed with err: %v, 重新推入jobQueue\n", err)
					jobQueue.ProcessJobQueue <- job
				}
				workerPool <- w.JobChannel
				// recieve quit event, stop worker
			case <-w.quit:
				return
			}
		}
	}(workerPool)
}
