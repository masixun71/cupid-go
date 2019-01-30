package jobQueue

type Job interface {
	Do() error
}

// define job channel
type JobChan chan Job

// define worker channer
type WorkerChan chan JobChan