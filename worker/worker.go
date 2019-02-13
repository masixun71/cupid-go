package worker

import (
	"../jobQueue"
	"fmt"
)

type Worker struct {
	JobChannel jobQueue.JobChan
	quit       chan bool
	workerPool jobQueue.WorkerChan
}

func NewWorker(workerPool jobQueue.WorkerChan) *Worker  {

	return &Worker{
		JobChannel:make(chan jobQueue.Job),
		quit: make(chan bool),
		workerPool: workerPool,
	}
}


func (w *Worker) Start() {
	w.workerPool <- w.JobChannel
	go func(workerPool jobQueue.WorkerChan) {
		for {
			select {
			case job := <-w.JobChannel:
				err := job.Do()
				job.IncrRetireTime()
				workerPool <- w.JobChannel
				if err != nil {
					fmt.Printf("excute job failed with err: %v, 推入失败队列\n", err)
					jobQueue.FailJobQueue <- job
				}
				// recieve quit event, stop worker
			case <-w.quit:
				return
			}
		}
	}(w.workerPool)
}
