package process

import (
	"../jobQueue"
	"../worker"
)

var Config = struct {
	WorkerNumber                      int    `default:"3"`
	LogDir                            string `default:"/apps/log/cupid"`
	FailureJobRetrySecond             int    `default:"10"`
	Src                               struct {
		Dsn                       string
		Table                     string
		ByColumn                  string
		Insert                    bool   `default:"true"`
		InsertIntervalMillisecond uint16 `default:"2000"`
		Update                    bool   `default:"true"`
		UpdateColumn              string
		UpdateIntervalMillisecond uint16 `default:"2000"`
		UpdateScanSecond          int    `default:"5"`
		UpdateTimeFormate         string `default:"2006-1-2 15:04:05"`
		CacheFilePath             string `default:"/tmp"`
		PushbearSendKey			  string `default:""`
	}
	Des []struct {
		Dsn                  string
		Table                string
		Columns              map[string]string
		ByColumn             string
		CallbackNotification struct {
			Url string
		}
	}
}{}

type Process struct {
	TaskDispatcher *worker.Dispatcher
	objectProcess  ObjectProcess
}

func NewProcess(objectProcess ObjectProcess, workerNumber int, processQueue jobQueue.JobChan) *Process {

	return &Process{
		TaskDispatcher: worker.NewDispatcher(workerNumber, processQueue),
		objectProcess:  objectProcess,
	}
}

func (p *Process) Run() {

	p.TaskDispatcher.Run()
	p.objectProcess.Init()

}


func (p *Process) Exit() {

	p.TaskDispatcher.Exit()
	p.objectProcess.Exit()
}
