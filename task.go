package main

import (
	"bytes"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/robfig/cron/v3" //定时任务
	log "github.com/sirupsen/logrus"
	"github.com/unknwon/goconfig"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//go mod init
var cfg *goconfig.ConfigFile

func main() {

	log.Info("读取配置文件")
	//读取配置文件
	cmds, _ := cfg.GetValue("java", "cmds") //读取单个值
	fmt.Println(cmds)
	cmds2, _ := cfg.GetValue("java", "cmds2") //读取单个值
	fmt.Println(cmds2)
	jarpath, _ := cfg.GetValue("java", "jarpath") //读取单个值
	fmt.Println(jarpath)
	starttime, _ := cfg.GetValue("schedule", "starttime") //读取单个值
	fmt.Println(starttime)
	endtime, _ := cfg.GetValue("schedule", "endtime") //读取单个值
	fmt.Println(endtime)
	startnow, _ := cfg.GetValue("java", "startnow") //读取单个值
	fmt.Println(startnow)

	log.WithFields(log.Fields{
		"cmds":      cmds,
		"cmds2":     cmds2,
		"jarpath":   jarpath,
		"starttime": starttime,
		"endtime":   endtime,
		"startnow":  startnow,
	}).Info("系统使用上述参数启动了")

	//是否需要立即启动
	if "yes" == startnow {

		log.Info("立即启动java程序...")
		timeStr := time.Now().Format("2006-01-02 15:04:05") //当前时间的字符串，2006-01-02 15:04:05据说是golang的诞生时间，固定写法
		log.Info("查找进程是否存在...")
		log.Info("当前时间：" + timeStr)
		//查找线程信息
		isExists, _, _ := isJarProcessExist(jarpath)
		log.Info("java程序是否已存在：" + strconv.FormatBool(isExists))
		if !isExists {
			go exeCmds(cmds, cmds2, jarpath)
		}
	}

	log.Debug("开始设定定时...")
	//设定定时任务
	//nyc, _ := time.LoadLocation("Asia/Shanghai")

	nyc, _ := time.LoadLocation("")
	var c = cron.New(cron.WithLocation(nyc)) //默认精确到分开始
	//var c = cron.New(cron.WithSeconds(), cron.WithLocation(nyc))
	//h := Hello{"I Love You!"}
	// 添加定时任务
	//c.AddJob("*/2 * * * * * ", h)
	// 添加定时任务
	c.AddFunc(starttime, func() {
		timeStr := time.Now().Format("2006-01-02 15:04:05") //当前时间的字符串，2006-01-02 15:04:05据说是golang的诞生时间，固定写法
		log.Info("启动定时开始任务...")
		log.Info("当前时间：" + timeStr)
		//查找线程信息
		isExists, _, _ := isJarProcessExist(jarpath)
		log.Info("java程序是否已存在：" + strconv.FormatBool(isExists))
		if !isExists {
			go exeCmds(cmds, cmds2, jarpath)
		}
	})
	c.AddFunc(endtime, func() {
		log.Info("启动定时结束任务...")
		timeStr := time.Now().Format("2006-01-02 15:04:05") //当前时间的字符串，2006-01-02 15:04:05据说是golang的诞生时间，固定写法
		log.Info("当前时间：" + timeStr)
		//查找线程信息

		log.Info("查找java程序是否已存在...")
		isExists, _, pid := isJarProcessExist(jarpath)

		log.Info("java程序是否已存在：" + strconv.FormatBool(isExists))
		if isExists {
			killProcess(strconv.Itoa(pid))
		}
	})

	//s, err := cron.Parse("*/3 * * * * *")
	//if err != nil {
	//	log.Println("Parse error")
	//}
	//h2 := Hello{"I Hate You!"}
	//c.Schedule(s, h2)
	// 其中任务
	c.Start()
	select {}
	defer c.Stop()
}

//初始化，自动调用
func init() {
	config, err := goconfig.LoadConfigFile("setting.conf") //加载配置文件

	//获取当前运行路径
	s, err := exec.LookPath(os.Args[0])
	i := strings.LastIndex(s, "\\")
	path := string(s[0 : i+1])
	fmt.Println(path)

	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("get config file error")
		os.Exit(-1)
	}
	cfg = config

	// 设置日志格式为json格式
	//设置输出样式，自带的只有两种样式logrus.JSONFormatter{}和logrus.TextFormatter{}
	//log.SetFormatter(&log.JSONFormatter{})
	//设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
	file, _ := os.OpenFile("schedule_task.log", os.O_CREATE|os.O_WRONLY, 0666)
	log.SetOutput(file)
	//设置最低loglevel
	log.SetLevel(log.InfoLevel)
	//日志分割hook
	log.AddHook(newLfsHook(100, "schedule_task.log"))
}

/**
针对的展名为.jar的java程序，使用jps进行查找线程号
*/
func isJarProcessExist(appName string) (bool, string, int) {
	cmd := exec.Command("cmd", "/C", "jps", "-l")
	log.Info("执行查找进程命令：" + cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("查找进程失败", "Command finished with error: %v", err)
	}
	//fmt.Printf("fields: %v\n", output)
	utf8, _ := GbkToUtf8(output)
	data := string(utf8)[:]
	fields := strings.Split(data, "\r\n")
	for _, value := range fields { //只需要值，不需要索
		if strings.Index(value, appName) > 0 {
			values := strings.Split(value, " ")
			int, err := strconv.Atoi(values[0])
			if err != nil {
				log.Fatal("查找进程失败", err.Error())
			}

			log.WithFields(log.Fields{
				"查找到的进程": appName,
				"pid":    int,
			}).Info("成功查线到进程")
			return true, appName, int
		}
	}
	return false, appName, -1
}

//杀掉某个线程
func killProcess(pid string) {
	cmd := exec.Command("cmd", "/C", "taskkill", "/f", "/t", "/pid", pid)
	log.Info("执行kill命令：" + cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("kill进程失败", err)
	}
	utf8, _ := GbkToUtf8(out)
	log.Info("kill进程结果：", string(utf8))
}

//执行某个命令
func exeCmds(text ...string) {

	cmd := exec.Command("cmd", "/C", text[0], text[1], text[2])

	//in := bytes.NewBuffer(nil)
	//cmd.Stdin = in //绑定输入
	var out bytes.Buffer
	cmd.Stdout = &out //绑定输出

	//go func() {
	//        // start stop restart	
	//        in.WriteString("nssm restart uplusSVCWB\n") //写入你的命令，可以有多行，"\n"表示回车
	//    }()
	err := cmd.Start()
	log.Info("执行启动命令：" + cmd.String())
	if err != nil {
		log.Fatal("启动失败", err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Command finished with error: ", err)
	}
	utf8, _ := GbkToUtf8(out.Bytes())
	log.Info(string(utf8))
}

//解决控制台中文乱码问题
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

//日志分割函数
func newLfsHook(maxRemainCnt uint, logName string) log.Hook {
	writer, err := rotatelogs.New(
		logName+".%Y%m%d%H",
		// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件
		rotatelogs.WithLinkName(logName),

		// WithRotationTime设置日志分割的时间，这里设置为一小时分割一次
		rotatelogs.WithRotationTime(time.Hour),

		// WithMaxAge和WithRotationCount二者只能设置一个，
		// WithMaxAge设置文件清理前的最长保存时间，
		// WithRotationCount设置文件清理前最多保存的个数。
		//rotatelogs.WithMaxAge(time.Hour*24),
		rotatelogs.WithRotationCount(maxRemainCnt),
	)

	if err != nil {
		log.Errorf("config local file system for logger error: %v", err)
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer,
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, &log.TextFormatter{DisableColors: true})

	return lfsHook
}
