# log

尚融log包, **兼容 logrus 的API**

## 该包对 logrus 做了哪些增强

该包在 logrus 的基础上, 增加以下几个功能:

- 打印日志时会带上代码行数

- 针对 sentry, graylog 和 日志切片功能的钩子做了封装

## Master分支状态

### 0.8-beta-3 (2017.9.1)

Fix #2

---

### v0.8-beta-2 (2017.8.31)

- 文件切片的hook不会按照Level级别将文件分开写,而是写在同一个日志文件里.

--- 

### v0.8-beta-1 (2017.8.31)

- 所有添加hook的方法不再需要传入Level的不定参数, 传入一个最低Level即可

---

### v0.7-beta-1 (2017.8.29)

Fix #1

---

### v0.7-beta (2017.8.29)

- 增加graylog钩子的添加方法:

---

### v0.6 (2017.8.25)

- 日志切片的hook增加了几个新的方法:

```go 
AddRotateHookWithFormatter(path string, maxAge, rotateTime time.Duration, format string, formatter Formatter, levels ...Level) error

AddRotateHookByDayWithFormatter(path string, maxAge, rotateDay int, formatter Formatter, levels ...Level) error

AddRotateHookByHourWithFormatter(path string, maxAge, rotateHour int, formatter Formatter, levels ...Level) error
```

同时增加了内置的Formatter ```PrefixedFormatter```, 该Formatter实现了如下格式的输出

```
time [Level][source] log content
```

使用例子:

```go

err := AddRotateHookByDayWithFormatter("log.log", 365, 1, log.PrefixedFormatter, log.InfoLevel) 
if err != nil {
    log.Error("fail to add a hook")
}
```

- 增加设置日志输出的方法:

```go
SetOutput(w io.Writer)
```

同时增加了输出到空设备的NullOutput变量, 它实现了 io.Writer 接口
若不想输出到磁盘文件外的其他任何地方, 调用如下代码即可:

```go
log.SetOutput(log.NullOutput)
```

- 日志文件不存在时,会自动创建所在的路径中包含的目录

---

### v0.5 (2017.8.24) 

基本功能测试版

---


```go
AddGrayLogHook(ip string, port int, extra map[string]interface{}, levels ...Level) error
AddAsyncGraylogHook(ip string, port int, extra map[string]interface{}, levels ...Level) error
GrayAsyncHookFlush()
```

使用例子:

```go

// gray 目前监听的地址: 192.168.1.101:12202/udp
err := log.AddGrayLogHook("192.168.1.101",12202, map[string]interface{}{"service":"my-service"}, log.InfoLevel)
if err != nil {
    log.Error("fail to add a hook")
}

```

```go 
err := log.AddAsyncGraylogHook("192.168.1.101",12202, map[string]interface{}{"service":"my-service"}, log.InfoLevel)
if err != nil {
    log.Error("fail to add a hook")
}

// 若使用异步方法记得在退出前清空缓冲区.
defer log.GrayAsyncHookFlush()
```


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

添加一个sentry hook

```go
func main() {

    // 在全局的logger中添加一个 sentry 的hook
    // 函数第一个参数为 sentry 的数据源地址
    // 该地址可以从 http://192.168.1.100:9000/sentry/ 获取
    err := log.AddSentryHook("http://ac5818c072e249ee9388d3610f641da8:815c23ee6cff4bc49b2b83db37144c98@192.168.1.100:9000/4", log.InfoLevel)
    if err != nil {
        log.Error("fail to add sentry hook to logrus")
    }

    // This log will sent to sentry
    log.Info("log")
}
```

添加一个异步的sentry hook

```go 
func main() {
    // 在全局的logger中添加一个 sentry 的异步hook
    err := log.AddAsyncSentryHook("http://ac5818c072e249ee9388d3610f641da8:815c23ee6cff4bc49b2b83db37144c98@192.168.1.100:9000/4", log.InfoLevel)
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
    // 将日志保存到文件log.log.info.[date], 并且日志的保存时间为永久, 日志的切分频率为1天, 触发的最小日志级别为 Info 
	err = log.AddRotateHookByDay("log.log", log.ForverDay, 1, log.InfoLevel)
    if err != nil {
        log.Error("fail to add rotate hook to logrus")
    }

     // 将日志保存到文件log.log.info.[date@time], 并且日志的保存时间为永久, 日志的切分频率为1小时, 触发的最小日志级别为 Info 
    err := log.AddRotateHookByHour("log.log", log.ForverHour, 1, log.InfoLevel)
    if err != nil {
        log.Error("fail to add rotate hook to logrus")
    }


    // 将日志保存到文件log.log.info.[date], 并且日志的保存时间为10分钟, 日志的切分频率为10分钟, 日志文件的命名格式为 "%Y-%m-%d@%H:%M" 触发的最小日志级别为 Info 
    err = log.AddRotateHook("log.log", time.Minute*10, time.Minute*10, "%Y-%m-%d@%H:%M", log.InfoLevel)
	if err != nil {
		log.Error("fail to add rotate hook to logrus")
	}

    log.Info("log")
    log.Error("log")
}
```

## 维护者 

目前该包的维护者是 IaaS 组的 @lwh 童鞋

如果发现该包的 bug 可以向他抛砖.希望增加一些你需要的功能可以提一个 PR 或者 issue 过来


## RoadMap

- [x] 将默认的全局logger对象的Level改为Debug (2017.8.24@14:20)
- [x] 日志文件不存在时自动创建,包括所在整个目录路径 (2017.8.24@15:25)
- [x] 支持设置日志格式
- [x] 能够针对日志文件的大小进行切片
