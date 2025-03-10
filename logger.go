package gamemitm

import (
	"fmt"
	"log"
)

// Level 日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger 日志接口
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	level Level
}

// ANSI颜色代码
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

// NewDefaultLogger 构造函数，初始化日志级别
func NewDefaultLogger(level ...int) *DefaultLogger {
	if len(level) > 0 {
		return &DefaultLogger{
			level: Level(level[0]),
		}
	}
	return &DefaultLogger{
		level: DEBUG,
	}
}

// Debug 输出调试日志
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		log.Printf(fmt.Sprintf("%s[DEBUG]%s %s", Blue, Reset, fmt.Sprintf(format, args...)))
	}
}

// Info 输出信息日志
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		log.Printf(fmt.Sprintf("%s[INFO]%s %s", Green, Reset, fmt.Sprintf(format, args...)))
	}
}

// Warn 输出警告日志
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		log.Printf(fmt.Sprintf("%s[WARN]%s %s", Yellow, Reset, fmt.Sprintf(format, args...)))
	}
}

// Error 输出错误日志
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		log.Printf(fmt.Sprintf("%s[ERROR]%s %s", Red, Reset, fmt.Sprintf(format, args...)))
	}
}

// Fatal 输出致命错误日志
func (l *DefaultLogger) Fatal(format string, args ...interface{}) {
	if l.level <= FATAL {
		log.Printf(fmt.Sprintf("%s[FATAL]%s %s", Red, Reset, fmt.Sprintf(format, args...)))
	}
}
