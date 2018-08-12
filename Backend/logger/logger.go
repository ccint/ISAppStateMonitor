package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"time"
	"path"
	"github.com/pkg/errors"
	"os"
)
var (
	Log *logrus.Logger
	logDir = "./log/"
	logFileName = "serverlog"
)

func Init() {
	Log = logrus.New()
	Log.SetLevel(logrus.InfoLevel)

	if _, err := os.Stat(logDir); err != nil && os.IsNotExist(err) {
		os.MkdirAll(logDir, 0755)
	}

	ConfigLocalFileSystemLogger(logDir, logFileName, time.Hour * 24 * 30, time.Hour * 24)
}

func ConfigLocalFileSystemLogger(logPath string, logFileName string, maxAge time.Duration, rotationTime time.Duration) {
	baseLogPath := path.Join(logPath, logFileName)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M%S",
		rotatelogs.WithLinkName(baseLogPath),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(rotationTime),
	)
	if err != nil {
		Log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	},&logrus.TextFormatter{DisableColors: true, TimestampFormat: "2006-01-02 15:04:05.000"})

	Log.AddHook(lfHook)
}
