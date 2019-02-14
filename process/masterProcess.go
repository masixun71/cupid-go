package process

import (
	. "../jobQueue"
	"go.uber.org/zap"
	"time"
)

type MasterProcess struct {
	quit chan bool
}

func (f *MasterProcess) Init() {

	f.quit = make(chan bool)
	InitDb(Config.WorkerNumber)
	f.initTimer()

}

func (f *MasterProcess) Exit() {
	f.quit <- true
}

func (p *MasterProcess) initTimer() {

	if Config.Src.Insert {
		idManage := &idManager{currentId: uint(StartId)}
		tInsert := time.NewTicker(time.Millisecond * time.Duration(Config.Src.InsertIntervalMillisecond))
		go func(t *time.Ticker, idManage *idManager) {
			for {
				select {
				case <-t.C:
					maxId := GetMaxId()
					if maxId > 0 {
						for idManage.currentId <= uint(maxId) {
							ProcessJobQueue <- &CompareCheckJob{Id: idManage.currentId}
							idManage.incrCurrentId()
						}
					}
				case <- p.quit:
					return
				}
				Logger.Info("当前insert定时扫描时间", zap.String("time", time.Now().Format("2006-1-2 15:04:05")))
			}
		}(tInsert, idManage)

	}
	if Config.Src.Update {

		tUpdate := time.NewTicker(time.Millisecond * time.Duration(Config.Src.UpdateIntervalMillisecond))
		go func(t *time.Ticker) {
			for {
				select {
				case <-t.C:
					updateIds := GetUpdateId()
					if len(updateIds) > 0 {
						for _, id := range updateIds {
							ProcessJobQueue <- &CompareCheckJob{Id: uint(id)}
						}
					}
				case <- p.quit:
					return
				}
				Logger.Info("当前update定时扫描时间", zap.String("time", time.Now().Format("2006-1-2 15:04:05")))
			}
		}(tUpdate)
	}

}
