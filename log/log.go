package log

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogsConfig struct {
	Debug *lumberjack.Logger
	Info  *lumberjack.Logger
	Err   *lumberjack.Logger
}

type Log struct {
}

func NewLog(config *LogsConfig) *zap.Logger {
	return zap.New(
		zapcore.NewTee(debug(config.Debug), info(config.Info), err(config.Err)),
		zap.AddCaller(),                   // 显示调用者信息（文件+行号）
		zap.AddStacktrace(zap.ErrorLevel), // 仅 Error 级别输出堆栈
	)
}

func newCore(config *lumberjack.Logger, level zapcore.Level) zapcore.Core {
	ensureLogDir(config.Filename)
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(config),
		&exactLevelEnabler{level: level},
	)
}

func debug(config *lumberjack.Logger) zapcore.Core {
	return newCore(config, zapcore.DebugLevel)
}

func info(config *lumberjack.Logger) zapcore.Core {
	return newCore(config, zapcore.InfoLevel)
}

func err(config *lumberjack.Logger) zapcore.Core {
	return newCore(config, zapcore.ErrorLevel)
}

func ensureLogDir(filename string) {
	dir := filename
	if i := strings.LastIndex(dir, "/"); i != -1 {
		dir = dir[:i]
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}
}

var encoderConfig = zapcore.EncoderConfig{
	TimeKey:        "time",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

// exactLevelEnabler 实现严格的级别过滤，只允许指定级别的日志
// 解决zap默认会包含所有更高级别日志的问题
type exactLevelEnabler struct {
	level zapcore.Level
}

func (e *exactLevelEnabler) Enabled(lvl zapcore.Level) bool {
	return lvl == e.level
}
