package process

import (
	. "../jobQueue"
	"encoding/gob"
	. "github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"go.uber.org/zap"
	"os"
	"time"
)

type FailProcess struct {
	quit          chan bool
	tickSyncQueue *SyncQueue
	TMSyncQueue   *SyncQueue // 10Minute queue
}

func NewFailProcess() *FailProcess {
	return &FailProcess{
		quit:          make(chan bool),
		tickSyncQueue: NewSyncQueue(),
		TMSyncQueue:   NewSyncQueue(),
	}
}

func (p *FailProcess) Run() {

	go func() {
		for {
			select {
			case job := <- FailJobQueue:
				p.tickSyncQueue.Push(job)
			case <-p.quit:
				return
			}
		}
	}()

	tick := time.NewTicker(time.Second * time.Duration(Config.FailureJobRetrySecond))
	go func(t *time.Ticker) {
		for {
			select {
			case <-t.C:
				Logger.Info("自定义时间间隔失败queue:开始处理")
				len := p.tickSyncQueue.Len()
				for i := 0; i < len; i++ {
					if job, ok := p.tickSyncQueue.TryPop(); ok {
						switch realJob := job.(type) {
						case Job:
							err := realJob.Do()
							realJob.IncrRetireTime()
							if err != nil {
								if realJob.GetRetireTime() <= 3 {
									Logger.Warn("自定义时间间隔失败queue:job执行失败, 重新推入自定义时间间隔失败queue", zap.String("job", realJob.String()))
									p.tickSyncQueue.Push(job)
								} else {
									p.TMSyncQueue.Push(job)
								}
							}
						}
					}
				}
				Logger.Info("自定义时间间隔失败queue:处理完成")
			case <- p.quit:
				return

			}
		}
	}(tick)

	tickTM := time.NewTicker(time.Minute * 10)
	go func(t *time.Ticker) {
		for {
			select {
			case <-t.C:
				Logger.Info("10Minutes,queue:开始处理")
				len := p.TMSyncQueue.Len()
				for i := 0; i < len; i++ {
					if job, ok := p.TMSyncQueue.TryPop(); ok {
						switch realJob := job.(type) {
						case Job:
							err := realJob.Do()
							realJob.IncrRetireTime()
							if err != nil {
								Logger.Warn("10Minutes,queue:job执行失败, 重新推入10Minutes,queue", zap.String("job", realJob.String()))
								p.TMSyncQueue.Push(job)
							}
						}
					}
				}
				Logger.Info("10Minutes,queue:处理完成")
			case <- p.quit:

				return
			}
		}
	}(tickTM)

	go func() {
		var list []Job
		file, err := os.Open(Config.Src.CacheFilePath + "/failJob.gob")
		if err != nil {
			Logger.Warn("打开缓存本地的失败job文件失败", zap.String("error", err.Error()))
			return
		}
		gob.Register(&CallbackJob{})
		gob.Register(&CompareCheckJob{})
		enc2 := gob.NewDecoder(file)
		err3 := enc2.Decode(&list)
		if err3 != nil {
			Logger.Warn("解码失败job文件:fail", zap.String("error", err.Error()))
			return
		}
		for _, job := range list {
			if job.GetRetireTime() <= 3 {
				p.tickSyncQueue.Push(job)
			} else {
				p.TMSyncQueue.Push(job)
			}
		}
	}()

}

func (p *FailProcess) Exit() {
	p.quit <- true

	gob.Register(&CallbackJob{})
	gob.Register(&CompareCheckJob{})
	file, err := os.Create(Config.Src.CacheFilePath + "/failJob.gob")
	if err != nil {
		Logger.Warn("创建缓存本地的失败job文件失败", zap.String("error", err.Error()))
	}
	list := make([]Job, 0)
	enc := gob.NewEncoder(file)
	for {
		if job, ok := p.tickSyncQueue.TryPop(); ok {
			switch realJob := job.(type) {
			case Job:
				list = append(list, realJob)
			}
		} else {
			break
		}
	}
	for {
		if job, ok := p.TMSyncQueue.TryPop(); ok {
			switch realJob := job.(type) {
			case Job:
				list = append(list, realJob)
			}
		} else {
			break
		}
	}
	err2 := enc.Encode(list)
	if err2 != nil {
		Logger.Warn("编码失败job文件:fail", zap.String("error", err2.Error()))
		return
	}
}
