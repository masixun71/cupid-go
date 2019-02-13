package process

import (
	"../jobQueue"
	"encoding/gob"
	"fmt"
	. "github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
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
			case job := <-jobQueue.FailJobQueue:
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
				fmt.Println("执行失败queue")
				len := p.tickSyncQueue.Len()
				for i := 0; i < len; i++ {
					if job, ok := p.tickSyncQueue.TryPop(); ok {
						switch realJob := job.(type) {
						case jobQueue.Job:
							err := realJob.Do()
							realJob.IncrRetireTime()
							if err != nil {
								fmt.Printf("excute job failed with err: %v, 推入失败queue\n", err)
								if realJob.GetRetireTime() <= 3 {
									p.tickSyncQueue.Push(job)
								} else {
									p.TMSyncQueue.Push(job)
								}
							}
						}
					}
				}
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
				fmt.Println("执行失败TMqueue")
				len := p.TMSyncQueue.Len()
				for i := 0; i < len; i++ {
					if job, ok := p.TMSyncQueue.TryPop(); ok {
						switch realJob := job.(type) {
						case jobQueue.Job:
							err := realJob.Do()
							realJob.IncrRetireTime()
							if err != nil {
								fmt.Printf("excute job failed with err: %v, 推入失败TMqueue\n", err)
								p.TMSyncQueue.Push(job)
							}
						}
					}
				}
			case <- p.quit:

				return
			}
		}
	}(tickTM)

	go func() {
		var list []jobQueue.Job
		file, err := os.Open(Config.Src.CacheFilePath + "/failJob.gob")
		if err != nil {
			fmt.Println(err)
			return
		}
		gob.Register(&CallbackJob{})
		gob.Register(&CompareCheckJob{})
		enc2 := gob.NewDecoder(file)
		err3 := enc2.Decode(&list)
		if err3 != nil {
			fmt.Println("decode", err3)
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
		fmt.Println(err)
	}
	list := make([]jobQueue.Job, 0)
	enc := gob.NewEncoder(file)
	for {
		if job, ok := p.tickSyncQueue.TryPop(); ok {
			switch realJob := job.(type) {
			case jobQueue.Job:
				list = append(list, realJob)
			}
		} else {
			break
		}
	}
	for {
		if job, ok := p.TMSyncQueue.TryPop(); ok {
			switch realJob := job.(type) {
			case jobQueue.Job:
				list = append(list, realJob)
			}
		} else {
			break
		}
	}
	err2 := enc.Encode(list)
	if err2 != nil {
		fmt.Println(err2)
		return
	}
}
