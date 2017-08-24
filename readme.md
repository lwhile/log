# log

尚融log包, **兼容 logrus 的API**

## 该包对 logrus 做了哪些增强

该包在 logrus 的基础上, 增加以下几个功能:

- 打印日志时会带上代码行数

- 针对 sentry 和 日志切片功能的钩子做了封装



## 如何使用

导入该包后就能直接使用一个全局的 logger 对象

```go
import (
    "git.gzsunrun.cn/sunrunlib/log"
)

func main() {
    log.Info("Hello world.")
}
```

当然你可以新建一个logger对象

```go

func main() {
    logger := log.NewLogger(os.Stdout)
}

```

添加一个不带标签的 sentry hook

```go
func main() {

    // 在全局的logger中添加一个 sentry 的hook
    // 函数第一个参数为 sentry 的数据源地址
    // 该地址可以从 http://192.168.1.100:9000/sentry/ 获取
    err := log.AddSentryHook("http://ac5818c072e249ee9388d3610f641da8:815c23ee6cff4bc49b2b83db37144c98@192.168.1.100:9000/4", log.InfoLevel,log.ErrorLevel)
    if err != nil {
        log.Error("fail to add sentry hook to logrus")
    }

    // This log will sent to sentry
    log.Info("log")
}
```


添加输出到文件的hook (带日志切片功能)

```go
func main() {
    // 将日志保存到文件log.log.info.[date], 并且日志的保存时间为永久, 日志的切分频率为1天, 触发的日志级别为 Info 和 Error
	err = log.AddRotateHookByDay("log.log", log.ForverDay, 1, log.InfoLevel, log.ErrorLevel)
    if err != nil {
        log.Error("fail to add rotate hook to logrus")
    }

     // 将日志保存到文件log.log.info.[date@time], 并且日志的保存时间为永久, 日志的切分频率为1小时, 触发的日志级别为 Info 和 Error
    err := log.AddRotateHookByHour("log.log", log.ForverHour, 1, log.InfoLevel, log.ErrorLevel)
    if err != nil {
        log.Error("fail to add rotate hook to logrus")
    }


    // 将日志保存到文件log.log.info.[date], 并且日志的保存时间为10分钟, 日志的切分频率为10分钟, 日志文件的命名格式为 "%Y-%m-%d@%H:%M" 触发的日志级别为 Info 和 Error
    err = log.AddRotateHook("log.log", time.Minute*10, time.Minute*10, "%Y-%m-%d@%H:%M", log.InfoLevel, log.ErrorLevel)
	if err != nil {
		log.Error("fail to add rotate hook to logrus")
	}

    log.Info("log")
    log.Error("log")
}
```

## 一些注意事项

- 添加hook的几个函数可以传入不定个log.Level类型的参数, hook只会对相应的log.Level级别生效.不会因为某些Level比另一些Level级别高也会生效.

- 添加 sentry 的 hook 会返回一个 error 变量, 这和添加日志 hook 不一样

- 目前只有我们公司内的环境有装 sentry, 所以在你的业务代码里需要判断配置文件是否有 sentry 数据源, 如果有才添加 sentry 的 hook



## 维护者 

目前该包的维护者是 IaaS 组的 @lwh 童鞋

如果该包有 bug 可以向他抛砖.希望增加一些你需要的功能可以提一个 PR 或者 issue 过来 