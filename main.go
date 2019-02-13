package main

import (
	"./jobQueue"
	"./process"
	"fmt"
	"github.com/jinzhu/configor"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {


	configor.Load(&process.Config, "/apps/conf/cupid/myConfigGo.json")

	jobQueue.ProcessJobQueue = make(jobQueue.JobChan, process.Config.WorkerNumber)
	jobQueue.CallbackJobQueue = make(jobQueue.JobChan, process.Config.WorkerNumber)
	jobQueue.FailJobQueue = make(jobQueue.JobChan, process.Config.WorkerNumber)

	callbackProcess := process.NewProcess(&process.CallbackProcess{}, 1, jobQueue.CallbackJobQueue)
	failProcess := process.NewFailProcess()
	callbackProcess.Run()
	failProcess.Run()

	masterProcess := process.NewProcess(&process.MasterProcess{}, process.Config.WorkerNumber, jobQueue.ProcessJobQueue)
	masterProcess.Run()

	InitSignal(masterProcess, callbackProcess, failProcess)

}

func InitSignal(masterProcess , callbackProcess *process.Process, failProcess *process.FailProcess) {

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg sync.WaitGroup) {
		for {
			c := make(chan os.Signal) //监听指定信号 ctrl+c kill
			signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
			s := <-c

			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				fmt.Println("exit", s)
				masterProcess.Exit()
				fmt.Println("masterProcess退出成功")
				callbackProcess.Exit()
				fmt.Println("callbackProcess退出成功")
				failProcess.Exit()
				fmt.Println("failProcess退出成功")
				wg.Done()
				os.Exit(0)
			case syscall.SIGUSR1: //kill -10 pid
				//SetLogLevel("DEBUG")  //切换日志级别DEBUG
				fmt.Println("usr1", s)
			case syscall.SIGUSR2: //kill -12 pid
				//SetLogLevel("ERROR")  //切换日志级别ERROR
				fmt.Println("usr2", s)
			default:
				fmt.Println("other", s)
			}
		}
	}(wg)

	wg.Wait()
}
