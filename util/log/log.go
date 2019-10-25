package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// error logger
var logger *zap.SugaredLogger

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func init() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, //这里可以指定颜色
		EncodeTime:     zapcore.ISO8601TimeEncoder,       // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 路径编码器
		//EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志级别
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(getLoggerLevel("debug")),     // 日志级别
		Development:      true,                                              // 开发模式，堆栈跟踪
		Encoding:         "console",                                         // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                     // 编码器配置
		InitialFields:    map[string]interface{}{"service": "configcenter"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// 构建日志
	oriLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(fmt.Sprintf("log init fail: %v", err))
	}
	logger = oriLogger.Sugar()
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	logger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	logger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}
