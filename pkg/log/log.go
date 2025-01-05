package log

import (
	"os"
	"time"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const LOGGER_KEY = "zapLogger"

type Logger struct {
	*zap.Logger
}

func NewLog(conf *config.Config) *Logger {
	return initZap(conf)
}

func initZap(conf *config.Config) *Logger {
	// 日志地址
	lp := conf.Log.LogFileName
	// 日志级别
	lv := conf.Log.LogLevel
	var level zapcore.Level
	// debug<info<warn<error<fatal<panic
	switch lv {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	hook := lumberjack.Logger{
		Filename:   lp,                  // 日志文件路径
		MaxSize:    conf.Log.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: conf.Log.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     conf.Log.MaxAge,     // 文件最多保存多少天
		Compress:   conf.Log.Compress,   // 是否压缩
	}

	var encoder zapcore.Encoder
	if conf.Log.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "Logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     timeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
			// EncodeCaller: zapcore.ShortCallerEncoder,
		})
	} else {
		encoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})
	}
	core := zapcore.NewCore(
		encoder, // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		level, // 日志级别
	)
	if conf.Env != "prod" {
		return &Logger{zap.New(core, zap.Development(), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))}
	}
	return &Logger{zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))}
}

// 自定义时间编码器
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// enc.AppendString(t.Format("2006-01-02 15:04:05"))
	enc.AppendString(t.Format("2006-01-02 15:04:05.000000000"))
}

// NewContext 给指定的context添加字段
func (l *Logger) NewContext(ctx *gin.Context, fields ...zapcore.Field) {
	ctx.Set(LOGGER_KEY, l.WithContext(ctx).With(fields...))
}

// WithContext 从指定的context返回一个zap实例
func (l *Logger) WithContext(ctx *gin.Context) *Logger {
	if ctx == nil {
		return l
	}
	zl, _ := ctx.Get(LOGGER_KEY)
	ctxLogger, ok := zl.(*zap.Logger)
	if ok {
		return &Logger{ctxLogger}
	}
	return l
}
