package lager

import (
	"log"
	"strings"

	paaslager "github.com/ServiceComb/paas-lager"
	"github.com/ServiceComb/paas-lager/third_party/forked/cloudfoundry/lager"
)

// Logger is the global variable for the object of lager.Logger
var Logger lager.Logger

// InitLager function will assign the lager.Logger value to Logger object which can be used further
func InitLager(log lager.Logger) {
	if log == nil {
		Logger = initialize("", "DEBUG", "", "size", true, 1, 10, 7)
	} else {
		Logger = log
	}
}

// Lager struct for logger parameters
type Lager struct {
	Writers        string `yaml:"writers"`
	LoggerLevel    string `yaml:"logger_level"`
	LoggerFile     string `yaml:"logger_file"`
	LogFormatText  bool   `yaml:"log_format_text"`
	RollingPolicy  string `yaml:"rollingPolicy"`
	LogRotateDate  int    `yaml:"log_rotate_date"`
	LogRotateSize  int    `yaml:"log_rotate_size"`
	LogBackupCount int    `yaml:"log_backup_count"`
}

// Initialize Build constructs a *Lager.Logger with the configured parameters.
func initialize(writers, loggerLevel, loggerFile, rollingPolicy string, logFormatText bool,
	LogRotateDate, LogRotateSize, LogBackupCount int) lager.Logger {
	lag := &Lager{
		Writers:        writers,
		LoggerLevel:    loggerLevel,
		LoggerFile:     loggerFile,
		LogFormatText:  logFormatText,
		RollingPolicy:  rollingPolicy,
		LogRotateDate:  LogRotateDate,
		LogRotateSize:  LogRotateSize,
		LogBackupCount: LogBackupCount,
	}

	log.Println("Enable log tool")
	return newLog(lag)
}

// newLog new log
func newLog(lag *Lager) lager.Logger {
	writers := strings.Split(strings.TrimSpace(lag.Writers), ",")
	if len(strings.TrimSpace(lag.Writers)) == 0 {
		writers = []string{"stdout"}
	}
	paaslager.Init(paaslager.Config{
		Writers:       writers,
		LoggerLevel:   lag.LoggerLevel,
		LogFormatText: lag.LogFormatText,
	})

	logger := paaslager.NewLogger(lag.LoggerFile)
	return logger
}
