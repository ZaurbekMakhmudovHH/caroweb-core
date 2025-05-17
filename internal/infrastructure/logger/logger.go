package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func Init(prod bool) {
	var cfg zap.Config

	if prod {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	// Отключаем stacktrace и caller
	cfg.EncoderConfig.StacktraceKey = "" // Убираем ключ stacktrace
	cfg.DisableCaller = true             // Не логируем caller (файл:строка)

	// Уровень логирования (можно настроить отдельно)
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	var err error
	Log, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}
