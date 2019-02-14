package main

import (
	. "./jobQueue"
	"./process"
	"fmt"
	"github.com/jinzhu/configor"
	"go.uber.org/zap"
	"gopkg.in/urfave/cli.v2"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
)

var (
	app = &cli.App{
		Name:    "cupid-go",
		Usage:   "简单快速的数据补偿工具，可监听insert和update数据",
		Version: "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "configPath",
				Aliases: []string{"c"},
				Value: "/tmp/cupid-go.json",
				Usage: "配置json，里面存储的是你对项目的配置和监听表的配置",
			},
			&cli.IntFlag{
				Name:  "startId",
				Aliases: []string{"s"},
				Value: 1,
				Usage: "起始的源table的id，没有的话从0开始，有的话从当前设置开始",
			},
		},
		//Commands: []*cli.Command{
		//	{
		//		Name:    "start",
		//		Aliases: []string{},
		//		Usage:   "开始我们的程序",
		//		Action: func(c *cli.Context) error {
		//			return nil
		//		},
		//
		//	},
		//},
	}
)

func init() {
	app.Action = rundeamon
	app.Before = func(ctx *cli.Context) error {
		fmt.Println("cupid-go 程序已开启")

		return nil
	}
	app.After = func(c *cli.Context) error {
		fmt.Println("cupid-go 程序成功关闭")
		return nil
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
}
func rundeamon(ctx *cli.Context) error {
	if args := ctx.Args(); args.Len() > 0 {
		return fmt.Errorf("invalid command: %q", args.Get(0))
	}
	configPath := ctx.String("configPath")
	startId := ctx.Int("startId")
	StartId = startId
	fmt.Println("使用的configPath", configPath)
	fmt.Println("设置的startId", startId)

	start(configPath, startId)

	return nil
}


func start(configPath string, startId int) {
	configor.Load(&process.Config, configPath)
	fmt.Println("日志输出路径", process.Config.LogDir)

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


func main() {

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}


}

func InitSignal(masterProcess, callbackProcess *process.Process, failProcess *process.FailProcess) {

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg sync.WaitGroup) {
		for {
			c := make(chan os.Signal) //监听指定信号 ctrl+c kill
			signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
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
				//case syscall.SIGUSR1: //kill -10 pid
				//	//SetLogLevel("DEBUG")  //切换日志级别DEBUG
				//	Logger.Info("进程收到SIGUSR1信号")
				//case syscall.SIGUSR2: //kill -12 pid
				//	//SetLogLevel("ERROR")  //切换日志级别ERROR
				//	Logger.Info("进程收到SIGUSR2信号")
			default:
				Logger.Info("进程收到其它信号", zap.String("signal", s.String()))
			}
		}
	}(wg)

	wg.Wait()
}
