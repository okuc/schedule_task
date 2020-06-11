 ## schedule_task

一个简单的定时启动、关闭java程序的小工具，在win10、win2008下测试通过。

**这是我第一个go语言的程序，高手请飘过~**

#### 编译
```
go get ./...
go build
```

#### release
可以直接下载编译好的 release 版本

提供 win64 平台的可执行文件

https://github.com/okuc/schedule_task/tree/master/release

#### 配置文件
`setting.conf`为本工具的配置文件，说明如下：
```
[java]
# java程序，需配置好java环境变量
cmds=java
# 指定运行方式
cmds2=-jar
# jar包路径
jarpath= D:/xxx/xxx/xxxx/XXXX.jar
# 是否启动时本工具时即启动java程序。
startnow=yes
[schedule]
# 定时启动时间，遵循cron规则，具体规则请百度
starttime=20/5 * * * *
# 定时启动时间，遵循cron规则，具体规则请百度
endtime=25/5 * * * *
```
#### 感谢
本工具使用到了以下第三方库，在此一并感谢。
**日志相关库**:
- [github.com/lestrrat-go/file-rotatelogs](github.com/lestrrat-go/file-rotatelogs)
- [github.com/rifflock/lfshook](github.com/rifflock/lfshook)
- [github.com/sirupsen/logrus](github.com/sirupsen/logrus)

**配置文件相关库**：
- [github.com/unknwon/goconfig](github.com/unknwon/goconfig)

**定时任务相关库**：
- [github.com/robfig/cron/v3](github.com/robfig/cron/v3)

最后，感谢google,创造了go语言！

#### LICENSE
Apache License 2.0