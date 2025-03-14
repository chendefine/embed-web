package embedweb

import (
	"io"
	"os"
	"path"
	"runtime"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)

	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
}

type logger struct {
	Logger
	level logrus.Level
}

func (l *logger) Debugf(format string, args ...any) {
	if l.level >= logrus.DebugLevel {
		l.Logger.Debugf(format, args...)
	}
}

func (l *logger) Infof(format string, args ...any) {
	if l.level >= logrus.InfoLevel {
		l.Logger.Infof(format, args...)
	}
}

func (l *logger) Warnf(format string, args ...any) {
	if l.level >= logrus.WarnLevel {
		l.Logger.Warnf(format, args...)
	}
}

func (l *logger) Errorf(format string, args ...any) {
	if l.level >= logrus.ErrorLevel {
		l.Logger.Errorf(format, args...)
	}
}

func (l *logger) Fatalf(format string, args ...any) {
	if l.level >= logrus.FatalLevel {
		l.Logger.Fatalf(format, args...)
	}
}

func (l *logger) Debug(args ...any) {
	if l.level >= logrus.DebugLevel {
		l.Logger.Debug(args...)
	}
}

func (l *logger) Info(args ...any) {
	if l.level >= logrus.InfoLevel {
		l.Logger.Info(args...)
	}
}

func (l *logger) Warn(args ...any) {
	if l.level >= logrus.WarnLevel {
		l.Logger.Warn(args...)
	}
}

func (l *logger) Error(args ...any) {
	if l.level >= logrus.ErrorLevel {
		l.Logger.Error(args...)
	}
}

func (l *logger) Fatal(args ...any) {
	if l.level >= logrus.FatalLevel {
		l.Logger.Fatal(args...)
	}
}

func (ew *EmbedWeb) initLogger() {
	lv := logrus.InfoLevel
	if ew.cfg != nil && ew.cfg.LogLevel != "" {
		lv, _ = logrus.ParseLevel(ew.cfg.LogLevel)
	}
	if embedLog == nil {
		embedLog = newEmbedLog()
	}
	if ew.log == nil {
		ew.log = &logger{Logger: embedLog, level: lv}
	} else {
		ew.log.level = lv
	}
}

func (ew *EmbedWeb) GetLogger() Logger {
	return ew.log
}

func (ew *EmbedWeb) SetLogger(log Logger) {
	if l, ok := log.(*logger); ok {
		log = l.Logger
	}
	ew.log.Logger = log
}

var (
	embedLog            Logger
	defaultLogFormatter = &nested.Formatter{TimestampFormat: time.DateTime, NoColors: true}
)

func newEmbedLog() Logger {
	log := logrus.New()
	logPath := path.Join(baseDirPath, embedLogFile)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("init embed log file error: %v", err)
		os.Exit(1)
	}
	log.SetOutput(makeLogWriter(logFile))
	log.SetFormatter(defaultLogFormatter)
	return log
}

func makeLogWriter(w io.Writer) io.Writer {
	switch runtime.GOOS {
	case "windows":
		return w
	default:
		return io.MultiWriter(w, os.Stdout)
	}
}
