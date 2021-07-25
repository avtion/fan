package main

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var log, _ = zap.NewProduction()

type cronLogger struct {
	ZapLogger *zap.Logger
}

var _ cron.Logger = (*cronLogger)(nil)

func (c *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	c.ZapLogger.Sugar().Infow(msg, keysAndValues...)
}

func (c *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	c.ZapLogger.With(zap.Error(err)).Sugar().Errorw(msg, keysAndValues...)
}

func getCronLogger() *cronLogger {
	return &cronLogger{ZapLogger: log}
}
