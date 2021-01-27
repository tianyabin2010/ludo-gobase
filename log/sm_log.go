package log

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	prefix        string
	fileWriter    = &logFileWriter{}
	nameFormat    = "2006-01-02"
	logTimeFormat = "2006-01-02 15:04:05.000"
	stdout        = false
)

func init() {
	ss := strings.Split(os.Args[0], "/")
	if "windows" == runtime.GOOS {
		ss = strings.Split(os.Args[0], "\\")
	}
	if runtime.GOOS != "linux" {
		stdout = true
	}
	prefix = ss[len(ss)-1]
	fileWriter.Init()
	zerolog.TimeFieldFormat = logTimeFormat
	log.Logger = log.Output(zerolog.SyncWriter(fileWriter))
	log.Error().Msg("log init successed!")
}

func createFile(now time.Time) (*os.File, error) {
	tstr := now.Format(nameFormat)
	name := fmt.Sprintf("%s%s_%s.log", "./", tstr, prefix)
	file, err := os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	return file, err
}

type logFileWriter struct {
	nextCreateTime int
	f              *os.File
}

func TomorrowStartTime(t time.Time) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05",
		t.AddDate(0, 0, 1).Format("2006-01-02")+" 00:00:00",
		time.Local)
}

func (l *logFileWriter) Init() {
	now := time.Now()
	f, err := createFile(now)
	if err != nil {
		os.Stderr.WriteString("create file error: " + err.Error())
	}
	l.f = f
	afterDay, err := TomorrowStartTime(now)
	if err == nil {
		l.nextCreateTime = int(afterDay.Unix())
	} else {
		l.nextCreateTime = int(now.AddDate(0, 0, 1).Unix())
	}
}

func (l *logFileWriter) Write(p []byte) (n int, err error) {
	l.CheckFileTime()
	if stdout {
		os.Stderr.Write(p)
	}
	if l.f != nil {
		return l.f.Write(p)
	}
	return 0, nil
}

func (l *logFileWriter) CheckFileTime() {
	now := time.Now()
	if int(now.Unix()) >= l.nextCreateTime {
		l.f.Sync()
		l.f.Close()
		f, err := createFile(now)
		if err != nil {
			os.Stderr.WriteString("create file error: " + err.Error())
		}
		l.f = f
		afterDay, err := TomorrowStartTime(now)
		if err == nil {
			l.nextCreateTime = int(afterDay.Unix())
		} else {
			l.nextCreateTime = int(now.AddDate(0, 0, 1).Unix())
		}
	}
}
