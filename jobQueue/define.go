package jobQueue

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Job interface {
	Do() error
	GetRetireTime() int
	IncrRetireTime()
	String() string
}

// define job channel
type JobChan chan Job

// define worker channer
type WorkerChan chan JobChan

var ProcessJobQueue JobChan

var CallbackJobQueue JobChan

var FailJobQueue JobChan

var Logger *zap.Logger

func InitLogger(logpath string, loglevel string) {
	hook := lumberjack.Logger{
		Filename:   logpath + "/cupid.log", // 日志文件路径
		MaxSize:    128,     // megabytes
		MaxBackups: 30,      // 最多保留30个备份
		MaxAge:     7,       // days
		Compress:   true,    // 是否压缩 disabled by default
	}

	w := zapcore.AddSync(&hook)

	// 设置日志级别,debug可以打印出info,debug,warn；info级别可以打印warn，info；warn只能打印warn
	// debug->info->warn->error
	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	// 时间格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		level,
	)

	Logger = zap.New(core)
	Logger.Info("DefaultLogger init success")
}
