package process

import (
	"../jobQueue"
	"../worker"
	"github.com/jinzhu/configor"
	"time"
)

var Config = struct {
	WorkerNumber                      int   `default:"3"`
	LogDir                            string `default:"/apps/log/cupid"`
	CallbackWorkerIntervalMillisecond uint16 `default:"1000"`
	TaskWorkerIntervalMillisecond     uint16 `default:"1000"`
	Src                               struct {
		Dsn                       string
		Table                     string
		ByColumn                  string
		Insert                    bool   `default:"true"`
		InsertIntervalMillisecond uint16 `default:"2000"`
		Update                    bool   `default:"true"`
		UpdateColumn              string
		UpdateIntervalMillisecond uint16 `default:"2000"`
		UpdateScanSecond          uint8  `default:"5"`
		UpdateTimeFormate         string `default:"Y-m-d H:i:s"`
		CacheFilePath             string `default:"/tmp"`
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
	TaskDispatcher     *worker.Dispatcher
}

func NewProcess(configPath string) *Process {
	configor.Load(&Config, configPath)

	return &Process{
		TaskDispatcher:     worker.NewDispatcher(Config.WorkerNumber),
	}
}

func (p *Process) Run() {

	jobQueue.ProcessJobQueue = make(jobQueue.JobChan)

	p.TaskDispatcher.Run()
	InitDb(Config.WorkerNumber)
	p.initTimer()

}

func (p *Process) initTimer() {

	if Config.Src.Insert {
		idManage := &idManager{currentId:1}
		tInsert := time.NewTicker(time.Millisecond * time.Duration(Config.Src.InsertIntervalMillisecond))
		go func(t *time.Ticker, idManage *idManager) {
			for {
				<-t.C
				maxId := GetMaxId()
				if maxId > 0 {
					for idManage.currentId <= uint(maxId) {

						p.TaskDispatcher.Consume(&CompareCheckJob{id:idManage.currentId})
						//fmt.Println("Insert:id", idManage.currentId)

						idManage.incrCurrentId()
					}
				}
			}
		}(tInsert, idManage)


	}
	if Config.Src.Update {
		tUpdate := time.NewTicker(time.Millisecond * time.Duration(Config.Src.UpdateIntervalMillisecond))
		go func(t *time.Ticker) {
			for {
				<-t.C
				//fmt.Println("get tUpdate", time.Now().Format("2006-1-2 15:04:05"))
			}
		}(tUpdate)
	}

}
