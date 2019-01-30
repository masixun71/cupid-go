package worker

import (
	"../jobQueue"
	"fmt"
)

type Dispatcher struct {
	workers      []*Worker
	quit         chan bool
	workerNumber int
	jobQueue     jobQueue.JobChan
	workerPool   jobQueue.WorkerChan
}

func NewDispatcher(workerNumber int) *Dispatcher {

	return &Dispatcher{
		workers:      make([]*Worker, 0),
		quit:         make(chan bool),
		workerNumber: workerNumber,
		jobQueue: make(jobQueue.JobChan),
		workerPool: make(jobQueue.WorkerChan, workerNumber),
	}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.workerNumber; i++ {
		worker := NewWorker()
		d.workers = append(d.workers, worker)
		worker.Start(d.workerPool)
	}

	go func() {
		for {
			select {
			case job := <-d.jobQueue:
					jobChan := <-d.workerPool
					jobChan <- job
				// stop dispatcher
			case <-d.quit:
				fmt.Print("error")
				return
			}
		}
	}()
}

func (d *Dispatcher) Consume(job jobQueue.Job)  {
	d.jobQueue <- job
}
