

![](https://img.shields.io/badge/version-v0.0.1-red.svg)
![](https://img.shields.io/badge/go-orange.svg)
![](https://img.shields.io/badge/多平台-blue.svg)

# cupid-go
Data synchronization compensation tool


# 简介
cupid-go是一个消息同步补偿工具，适用于在`canal`这类的实时同步数据中间件之外的同步补偿工具，也适用于实时缓存这类的及时更新工具，对数据同步进行双保险，一般`canal`的延迟在毫秒左右，`cupid`建议设置在秒左右，做补偿专用，`canal`的消费端嵌入业务代码可以更方便开发和消费，`cupid`作为更通用的补偿方案,所以建议不要嵌入业务代码，补偿机制采用http回调来保证，失败会重试直到成功，若回调失败会通知（目前采用`pushbear`微信即时通知,简单即时）。


# 环境要求


- 编译源代码建议`go`版本在1.10以上



# 安装步骤

## 1.可选择编译代码或直接下载执行文件
```go
若采用编译代码，执行
go build  cupid.go
```
或直接去以下链接下载

https://github.com/masixun71/cupid-go/releases



## 2.填写一个我们的配置文件（任意目录，要求json）
#### 配置详情 (可参考testConfig.json)
```php
    {
    "workerNumber": 3, //最低值3，前2个进程是manager和callback进程，之后的才是处理进程
	"logDir": "/tmp", //日志目录，进程日志会打印到该目录下
	"callbackWorkerIntervalMillisecond": 1000, //回调进程间隔多长时间处理一次
    "taskWorkerIntervalMillisecond": 1000, //task进程间隔多长时间处理一次
	"failureJobRetrySecond": 10,  //失败队列重试间隔
    "src": {
        "dsn": "user:password@tcp(127.0.0.1:3306)/db?charset=utf8",//数据库dsn配置
        "table": "user",//我们关注的数据库表
        "byColumn": "number",//通过该字段对应des的byColumn，进行比对
        "insert": true, //是否关注insert的数据
        "insertIntervalMillisecond": 2000, //检查insert更新的间隔时间，单位毫秒
        "update": true, //是否关注update数据，若为false，则下面update开头的字段可以不用
        "updateColumn": "update_time", //update为true时必填，更新的字段名，需要添加索引，不然会扫描全表
        "updateIntervalMillisecond": 2000, //update为true时必填,检查update更新的间隔时间，单位毫秒
        "updateScanSecond": 5, //update为true时必填,获取数据的时间间隔，当前时间减去updateScanSecond设的时间为开始时间，当前时间为结束时间
        "updateTimeFormate": "Y-m-d H:i:s", //update为true时必填,数据库里数据更新字段的时间格式
        "cacheFilePath": "/tmp", //若进程有异常退出或者重启，会把当前的遍历信息记录到缓存文件中，重启时直接读取缓存文件
        "pushbearSendKey": "9724-73bdacb319007f53f83d0123213b4ec964"//若需要pushbear推送微信消息，在这填写
    },
    "des": [//des是一个数组，意味着我们可同时比对多个数据表
        {
            "dsn": "user:password@tcp(127.0.0.1:3306)/db2?charset=utf8",//数据库dsn配置
            "table": "user",//我们同步的数据库表
            "columns": { //关注和同步表的字段对应关系
                "number": "number",
                "name": "name",
                "avatar": "avatar"
            },
            "byColumn": "number",//与src的byColumn相呼应，形成关联关系来比对
            "callbackNotification": {
                "url" : "127.0.0.1:20000/test/callback"//当数据不同步时的回调地址
            }

        }
    ]
}
```

# 启动项目
```php
./cupid -c /tmp/myConfig.json -s 1
```

- **-c** 或者`--configPath`（必须）: config配置的路径，需要绝对路径 ，对config配置有问题可以看文档或者testConfig.json
- **-s**或者`--startId`(非必须): 起始的数据库表id, 优先级： shell命令传入的start_id > 默认值1

# 帮助

```php
./cupid -h
```

# pushbear

pushbear是一个基于微信模板的一对多消息送达服务，使用简单,高效,只需要申请一个key即可。

[pushbear官网](http://pushbear.ftqq.com/admin/#/)



# 回调形式

回调采用的是`POST`形式，Content-Type 为 `application/json`

```php
{
	"type" : 1, 
    "srcColumn" : {
    	"column1": "1",
    	"column2": "2",
    	"column3": null
    }
}
```

- type, 指的是变更类型，insert是1,update是2
- srcColumn, 指的是源数据列，会把整个源数据传给你, 需要注意的一点 **srcColumn传的值若有的都是string类型，没有则是null**

### 回调接口时的超时时间为3秒

### 回调失败后会推入失败队列，等待重新回调


# supervisord

该项目比较适合搭配supervisord使用,基本上配置文件应该如下

```php
[program:cupid.synchronization-compensation-tool]
directory=/tmp
command=/tmp/cupid -c /tmp/myConfig.json -s 1
numprocs=1
autorestart=true
stopsignal=TERM
stopwaitsecs=2
killasgroup=true
user=nobody
stdout_logfile=/tmp/cupid.synchronization-compensation-tool.log
redirect_stderr=true
loglevel=debug
```

如果你不用supervisord, 建议重启进程时使用`kill -s SIGTERM $master_pid`



# 进程正常退出或异常退出

会将所有失败队列里的任务持久化到本地文件中，以便下次启动时重新推入失败队列重试，持久化配置参见`config.json`



# 注意事项

- 回调接口时，传的srcColumn, 指的是源数据列，会把整个源数据传给你, 需要注意的一点 **srcColumn传的值若有的都是string类型，没有则是null**
- 所有在自定义时间间隔失败重试队列超过三次仍未成功的任务会推送到10分钟重试队列里



# todo

- pushbear 正在接入
- 处理id还未打印，内存损耗未统计