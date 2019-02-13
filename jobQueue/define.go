package jobQueue

type Job interface {
	Do() error
	GetRetireTime() int
	IncrRetireTime()
}

// define job channel
type JobChan chan Job

// define worker channer
type WorkerChan chan JobChan


var ProcessJobQueue JobChan

var CallbackJobQueue JobChan

var FailJobQueue JobChan