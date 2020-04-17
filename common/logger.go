package common

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
	"time"
)

var Log = logrus.New()

const(
	filePath string = "simple-printer.log"
)

func InitLog() {
	// 设置日志格式为json
	Log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置将日志输出到标准输出（黑夜的输出为stderr，标准错误）
	// 日志消息输出可以是做生意的 io.writer 类型
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		//TODO 异常处理
	}
	Log.SetOutput(file)

	// 设置日志级别为 warn 以上
	Log.SetLevel(logrus.InfoLevel)

	Log.AddHook(&LineHook{})
	Log.AddHook(newLfsHook())
}

// line number hook for log the call context,
type LineHook struct {
	Field  string
	// skip为遍历调用栈开始的索引位置
	Skip   int
	levels []logrus.Level
}

// Levels implement levels
func (hook LineHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire implement fire
func (hook LineHook) Fire(entry *logrus.Entry) error {
	entry.Data[hook.Field] = findCaller(hook.Skip)
	return nil
}

func findCaller(skip int) string {
	file := ""
	line := 0
	var pc uintptr
	// 遍历调用栈的最大索引为第11层.
	for i := 0; i < 11; i++ {
		file, line, pc = getCaller(skip + i)
		// 过滤掉所有logrus包，即可得到生成代码信息
		if !strings.HasPrefix(file, "logrus") {
			break
		}
	}

	fullFnName := runtime.FuncForPC(pc)

	fnName := ""
	if fullFnName != nil {
		fnNameStr := fullFnName.Name()
		// 取得函数名
		parts := strings.Split(fnNameStr, ".")
		fnName = parts[len(parts)-1]
	}

	return fmt.Sprintf("%s:%d:%s()", file, line, fnName)
}

func getCaller(skip int) (string, int, uintptr) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0, pc
	}
	n := 0

	// 获取包名
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return file, line, pc
}

func newLfsHook() logrus.Hook {
	writer, err := rotatelogs.New(
		"./logs/" + filePath+".%Y%m%d%H",
		// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件
		rotatelogs.WithLinkName(filePath),

		// WithRotationTime设置日志分割的时间，这里设置为一小时分割一次
		rotatelogs.WithRotationTime(time.Hour),

		// WithMaxAge和WithRotationCount二者只能设置一个，
		// WithMaxAge设置文件清理前的最长保存时间，
		// WithRotationCount设置文件清理前最多保存的个数。
		rotatelogs.WithMaxAge(time.Hour*24),
		//rotatelogs.WithRotationCount(maxRemainCnt),
	)

	if err != nil {
		logrus.Errorf("config local file system for logger error: %v", err)
	}

	logrus.SetLevel(logrus.InfoLevel)

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &logrus.TextFormatter{DisableColors: true})

	return lfsHook
}