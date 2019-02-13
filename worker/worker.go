package worker

import (
	. "../jobQueue"
	"go.uber.org/zap"
)

type Worker struct {
	JobChannel JobChan
	quit       chan bool
	workerPool WorkerChan
}

func NewWorker(workerPool WorkerChan) *Worker  {

	return &Worker{
		JobChannel:make(chan Job),
		quit: make(chan bool),
		workerPool: workerPool,
	}
}


func (w *Worker) Start() {
	w.workerPool <- w.JobChannel
	go func(workerPool WorkerChan) {
		for {
			select {
			case job := <-w.JobChannel:
				err := job.Do()
				job.IncrRetireTime()
				workerPool <- w.JobChannel
				if err != nil {
					Logger.Warn("worker执行job执行失败, 推入失败队列", zap.String("job", job.String()), zap.String("error", err.Error()))
					FailJobQueue <- job
				}
				// recieve quit event, stop worker
			case <-w.quit:
				return
			}
		}
	}(w.workerPool)
}
