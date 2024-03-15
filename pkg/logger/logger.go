package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Provide(New)

type Logger interface {
	Logger() *zap.SugaredLogger
}

type logger struct {
	logger *zap.SugaredLogger
}

type Params struct {
	fx.In
}

func New(p Params) Logger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"allie.log",
	}
	cfg.DisableStacktrace = true
	l, _ := cfg.Build()
	sl := l.Sugar()

	return &logger{
		logger: sl,
	}
}

func (l *logger) Logger() *zap.SugaredLogger {
	return l.logger
}
