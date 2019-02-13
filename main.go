package main

import (
	. "./jobQueue"
	"./process"
	"github.com/jinzhu/configor"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	configor.Load(&process.Config, "/apps/conf/cupid/myConfigGo.json")

	InitLogger(process.Config.LogDir, "debug")
	ProcessJobQueue = make(JobChan, process.Config.WorkerNumber)
	CallbackJobQueue = make(JobChan, process.Config.WorkerNumber)
	FailJobQueue = make(JobChan, process.Config.WorkerNumber)

	callbackProcess := process.NewProcess(&process.CallbackProcess{}, 1, CallbackJobQueue)
	failProcess := process.NewFailProcess()
	callbackProcess.Run()
	failProcess.Run()

	masterProcess := process.NewProcess(&process.MasterProcess{}, process.Config.WorkerNumber, ProcessJobQueue)
	masterProcess.Run()

	InitSignal(masterProcess, callbackProcess, failProcess)

}

func InitSignal(masterProcess, callbackProcess *process.Process, failProcess *process.FailProcess) {

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg sync.WaitGroup) {
		for {
			c := make(chan os.Signal) //监听指定信号 ctrl+c kill
			signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
			s := <-c

			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				Logger.Info("进程收到退出信号", zap.String("signal", s.String()))
				masterProcess.Exit()
				Logger.Info("masterProcess退出成功")
				callbackProcess.Exit()
				Logger.Info("callbackProcess退出成功")
				failProcess.Exit()
				Logger.Info("failProcess退出成功")
				wg.Done()
				os.Exit(0)
			case syscall.SIGUSR1: //kill -10 pid
				//SetLogLevel("DEBUG")  //切换日志级别DEBUG
				Logger.Info("进程收到SIGUSR1信号")
			case syscall.SIGUSR2: //kill -12 pid
				//SetLogLevel("ERROR")  //切换日志级别ERROR
				Logger.Info("进程收到SIGUSR2信号")
			default:
				Logger.Info("进程收到其它信号", zap.String("signal", s.String()))
			}
		}
	}(wg)

	wg.Wait()
}
