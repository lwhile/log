# log

尚融log库, **兼容 logrus 的API**

## Example


- 添加 sentry 的 hook

```go

import (
    "git.gzsunrun.cn/sunrunlib/log"
)

func main() {
    err := log.AddSentryHook("http://ac5818c072e249ee9388d3610f641da8:815c23ee6cff4bc49b2b83db37144c98@192.168.1.100:9000/4", log.InfoLevel,log.ErrorLevel)
    if err != nil {
        log.Error("fail to add sentry hook to logrus")
    }

    // This log will sent to sentry
    log.Info("log")
}

```  

- 添加日志文件切片的 hook

```go 

import (
    "git.gzsunrun.cn/sunrunlib/log"
)

func main() {
    // 将日志保存到文件log.log, 并且日志的保存时间为永久, 日志的切分频率为1小时, 触发的日志级别为 Info 和 Error
    err := log.AddRotateHookByHour("log.log", log.ForverHour, 1, log.InfoLevel, log.ErrorLevel)
    if err != nil {
        log.Error("fail to add rotate hook to logrus")
    }

    // 将日志保存到文件log.log, 并且日志的保存时间为永久, 日志的切分频率为1天, 触发的日志级别为 Info 和 Error
	err = log.AddRotateHookByDay("log.log", log.ForverDay, 1, log.InfoLevel, log.ErrorLevel)
    if err != nil {
        log.Error("fail to add rotate hook to logrus")
    }

    // 将日志保存到文件log.log, 并且日志的保存时间为10分钟, 日志的切分频率为10分钟, 触发的日志级别为 Info 和 Error
    err = log.AddRotateHook("log.log", time.Minute*10, time.Minute*10, "%Y-%m-%d@%H:%M", log.InfoLevel, log.ErrorLevel)
	if err != nil {
		log.Error("fail to add rotate hook to logrus")
	}

    log.Info("log")
    log.Error("log")
}

```

