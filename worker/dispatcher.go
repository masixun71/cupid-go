package worker

import (
	"../jobQueue"
)

type Dispatcher struct {
	workers      []*Worker
	quit         chan bool
	workerNumber int
	workerPool   jobQueue.WorkerChan
	processQueue  jobQueue.JobChan
}

func NewDispatcher(workerNumber int, processQueue  jobQueue.JobChan) *Dispatcher {


	return &Dispatcher{
		workers:      make([]*Worker, 0),
		quit:         make(chan bool),
		workerNumber: workerNumber,
		workerPool: make(jobQueue.WorkerChan, workerNumber),
		processQueue: processQueue,
	}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.workerNumber; i++ {
		worker := NewWorker(d.workerPool)
		d.workers = append(d.workers, worker)
		worker.Start()
	}

	go func() {
		for {
			select {
			case job := <- d.processQueue:
					jobChan := <-d.workerPool
					jobChan <- job
				// stop dispatcher
			case <-d.quit:
				return
			}
		}
	}()
}


func (d *Dispatcher) Exit() {
	d.quit <- true
	for _, worker := range d.workers {
		worker.quit <- true
	}


}
