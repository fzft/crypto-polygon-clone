package log

import (
	"bytes"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	Log = log.New()
)

func ParseLevel(lvl string) (log.Level, error) {
	return log.ParseLevel(lvl)
}

type TextFormatter struct {
	log.TextFormatter
}

func (f *TextFormatter) Format(entry *log.Entry) ([]byte, error) {
	data, err := f.TextFormatter.Format(entry)
	if err != nil {
		return data, err
	}

	b := new(bytes.Buffer)
	b.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	b.WriteByte(' ')
	b.Write(data)
	return b.Bytes(), nil
}

func SetupLogger() {
	logLevel := viper.GetString("logging.level")
	logFile := viper.GetString("logging.file")

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Log.Fatalf("error opening file: %v", err)
		}
		Log.SetOutput(f)
	}

	lvl, _ := ParseLevel(logLevel)
	Log.SetLevel(lvl)

	Log.SetFormatter(&TextFormatter{
		log.TextFormatter{
			DisableLevelTruncation: true,
			ForceColors:            true,
			DisableTimestamp:       true,
		},
	})
}

func Printf(msg string, v ...interface{}) {
	Log.Printf(msg, v...)
}

func Println(msg ...interface{}) {
	Log.Println(msg...)
}

func Fatalf(msg string, v ...interface{}) {
	Log.Fatalf(msg, v...)
}

func Fatal(msg ...interface{}) {
	Log.Fatal(msg...)
}

func Info(msg ...interface{}) {
	Log.Info(msg...)
}

func Infof(fmt string, args ...interface{}) {
	Log.Infof(fmt, args...)
}

func Debug(msg ...interface{}) {
	Log.Debug(msg...)
}

func Debugf(fmt string, args ...interface{}) {
	Log.Debugf(fmt, args...)
}

func Warn(msg ...interface{}) {
	Log.Warn(msg...)
}

func Warnf(fmt string, args ...interface{}) {
	Log.Warnf(fmt, args...)
}

func Error(msg ...interface{}) {
	Log.Error(msg...)
}

func Errorf(fmt string, args ...interface{}) {
	Log.Errorf(fmt, args...)
}

func Trace(msg ...interface{}) {
	Log.Trace(msg...)
}

func Tracef(fmt string, args ...interface{}) {
	Log.Tracef(fmt, args...)
}

func GetLevel() log.Level {
	return Log.GetLevel()
}
