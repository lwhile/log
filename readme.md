# log

尚融log库, **兼容 logrus 的API**

## Example


- 添加 sentry 的 hook

```go

import (
    "git.gzsunrun.cn/sunruniaas/sunruniaas-utils/log"
)

func main() {
    err := log.AddSentryHook("http://ac5818c072e249ee9388d3610f641da8:815c23ee6cff4bc49b2b83db37144c98@192.168.1.100:9000/4", log.InfoLevel)
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
    "git.gzsunrun.cn/sunruniaas/sunruniaas-utils/log"
)

func main() {
    
    // 添加以小时为单位进行日志切分的hook
    // 倒数第二个参数为日志的保留时间,单位为小时
    // 倒数第一个参数为日志的分割周期,单位为小时
    // log.ForverHour 表示保留所有日志
    err := log.AddRotateHookByHour("log.log", log.InfoLevel, log.ForverHour, 1)
    if err != nil {
        log.Error("fail to add rotate hook to logrus")
    }

    // 添加以小时为单位进行日志切分的hook
    // 倒数第二个参数为日志的保留时间,单位为天
    // 倒数第一个参数为日志的分割周期,单位为天
    // log.ForverDay 表示保留所有日志
	err = log.AddRotateHookByDay("log.log", log.InfoLevel, log.ForverDay, 1)
    if err != nil {
        log.Error("fail to add rotate hook to logrus")
    }

    log.Info("log")
}

```

