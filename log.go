package goweb

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

func debug(config *lumberjack.Logger) zapcore.Core {
	// 获取文件目录
	dir := config.Filename
	if i := strings.LastIndex(dir, "/"); i != -1 {
		dir = dir[:i]
	}
	// 确保日志目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}

	// 日志编码配置（JSON 格式）
	encoderConfig := zapcore.EncoderConfig{
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

	// 文件 Core：仅处理 Debug 级别
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(config),
		&exactLevelEnabler{level: zapcore.DebugLevel},
	)
}

func info(config *lumberjack.Logger) zapcore.Core {
	// 获取文件目录
	dir := config.Filename
	if i := strings.LastIndex(dir, "/"); i != -1 {
		dir = dir[:i]
	}
	// 确保日志目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}

	// 日志编码配置（JSON 格式）
	encoderConfig := zapcore.EncoderConfig{
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

	// 文件 Core：仅处理 Info 级别
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(config),
		&exactLevelEnabler{level: zapcore.InfoLevel},
	)
}

func err(config *lumberjack.Logger) zapcore.Core {
	// 获取文件目录
	dir := config.Filename
	if i := strings.LastIndex(dir, "/"); i != -1 {
		dir = dir[:i]
	}
	// 确保日志目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}

	// 日志编码配置（JSON 格式）
	encoderConfig := zapcore.EncoderConfig{
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

	// 文件 Core：仅处理 Error 级别
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(config),
		&exactLevelEnabler{level: zapcore.ErrorLevel},
	)
}

// exactLevelEnabler 实现严格的级别过滤，只允许指定级别的日志
// 解决zap默认会包含所有更高级别日志的问题
type exactLevelEnabler struct {
	level zapcore.Level
}

func (e *exactLevelEnabler) Enabled(lvl zapcore.Level) bool {
	return lvl == e.level
}
