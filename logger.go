package eweb

import (
	"io"
	"os"
	"path"
	"runtime"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var (
	defaultLogFormatter = &nested.Formatter{TimestampFormat: time.DateTime, NoColors: true}
)

func newEmbedLog() *logrus.Logger {
	log := logrus.New()
	logPath := path.Join(baseDirPath, embedLogFile)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("init embed log file error: %v", err)
	}
	log.SetOutput(makeLogWriter(logFile))
	log.SetFormatter(defaultLogFormatter)
	return log
}

func (ew *EmbedWeb) initAppLog() {
	ew.log = logrus.New()

	logLevel, err := logrus.ParseLevel(ew.config.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	ew.log.SetLevel(logLevel)

	logPath := path.Join(baseDirPath, appLogFile)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		ew.embedLog.Errorf("init app log file error: %v", err)
	}
	ew.log.SetOutput(makeLogWriter(logFile))
	ew.log.SetFormatter(defaultLogFormatter)
}

func makeLogWriter(w io.Writer) io.Writer {
	switch runtime.GOOS {
	case "windows":
		return w
	default:
		return io.MultiWriter(w, os.Stdout)
	}
}
