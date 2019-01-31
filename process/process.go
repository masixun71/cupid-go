package process

import (
	"../jobQueue"
	"../worker"
	"fmt"
	"github.com/jinzhu/configor"
	"time"
)

var Config = struct {
	WorkerNumber                      int   `default:"3"`
	LogDir                            string `default:"/apps/log/cupid"`
	CallbackWorkerIntervalMillisecond uint16 `default:"1000"`
	TaskWorkerIntervalMillisecond     uint16 `default:"1000"`
	FailureJobRetrySecond			  int	 `default:"10"`
	Src                               struct {
		Dsn                       string
		Table                     string
		ByColumn                  string
		Insert                    bool   `default:"true"`
		InsertIntervalMillisecond uint16 `default:"2000"`
		Update                    bool   `default:"true"`
		UpdateColumn              string
		UpdateIntervalMillisecond uint16 `default:"2000"`
		UpdateScanSecond          int  `default:"5"`
		UpdateTimeFormate         string `default:"2006-1-2 15:04:05"`
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

	jobQueue.ProcessJobQueue = make(jobQueue.JobChan, 1000)
	jobQueue.FailJobQueue = make(jobQueue.JobChan, Config.FailureJobRetrySecond * 100)

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
						idManage.incrCurrentId()
					}
				}
				fmt.Println("get tInsert", time.Now().Format("2006-1-2 15:04:05"))
			}
		}(tInsert, idManage)


	}
	if Config.Src.Update {

		tUpdate := time.NewTicker(time.Millisecond * time.Duration(Config.Src.UpdateIntervalMillisecond))
		go func(t *time.Ticker) {
			for {
				<-t.C
				updateIds := GetUpdateId()
				if len(updateIds) > 0 {
					for _, id := range updateIds {
						p.TaskDispatcher.Consume(&CompareCheckJob{id:uint(id)})
					}
				}


				fmt.Println("get tUpdate", time.Now().Format("2006-1-2 15:04:05"))
			}
		}(tUpdate)
	}

	tFail := time.NewTicker(time.Second * time.Duration(Config.FailureJobRetrySecond))
	go func(t *time.Ticker) {
		for {
			<- t.C
			fmt.Println("失败队列开始处理")
			isLoop := true
			for {
				if !isLoop {
					break
				}
				select {

				case job :=<- jobQueue.FailJobQueue:
					fmt.Println("从失败队列取出job发送到procesJob", job)
					p.TaskDispatcher.Consume( job)
					fmt.Println("从失败队列取出job发送到procesJob完毕", job)
				default:
					fmt.Println("从失败队列没有取出job发送到procesJob")
					isLoop = false
				}
			}

			fmt.Println("失败队列处理完毕")
		}
	}(tFail)


}
