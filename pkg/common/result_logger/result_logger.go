package result_logger

// zap + lumberjack log rotator,
// directly using https://juejin.cn/post/6844903727464185869.
// many thanks to the author.

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

var logLevel = zap.NewAtomicLevel()

func InitLogger(folder, file, level string) {
	SetLevel(configLevel[level])
	filePath := getFilePath(folder, file)

	fmt.Println("[INFO] logger init filePath: ", filePath)

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    1024, //MB
		MaxBackups: 7,
		MaxAge:     7, //days
		LocalTime:  true,
		Compress:   false,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			w,
			logLevel,
		),
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stdout),
			logLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	log = logger.Sugar()
}

type Level int8

const (
	DebugLevel Level = iota - 1

	InfoLevel

	WarnLevel

	ErrorLevel

	DPanicLevel

	PanicLevel

	FatalLevel
)

var configLevel = map[string]Level{
	"debug": DebugLevel,
	"info":  InfoLevel,
	"warn":  WarnLevel,
	"error": ErrorLevel,
	"panic": PanicLevel,
	"fatal": FatalLevel,
}

func SetLevel(level Level) {
	logLevel.SetLevel(zapcore.Level(level))
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Info(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func getFilePath(folder, file string) string {
	logfile := getCurrentDirectory() + "/" + folder + "/" + file
	return logfile
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	log.Debugf(template, args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(template string, args ...interface{}) {
	log.Infof(template, args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	log.Warnf(template, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	log.Errorf(template, args...)
}

func Panic(args ...interface{}) {
	log.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	log.Panicf(template, args...)
}

func With(args ...interface{}) *zap.SugaredLogger {
	return log.With(args...)
}
