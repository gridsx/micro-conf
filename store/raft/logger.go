package raft

import (
	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"

	"io"
	"log"
)

type CustomLogger struct {
	logger *logrus.Logger
	name   string
}

func NewCustomLogger(l *logrus.Logger, name string) *CustomLogger {
	return &CustomLogger{
		logger: l,
		name:   name,
	}
}

func (c *CustomLogger) Log(level hclog.Level, msg string, args ...interface{}) {
	c.logger.Log(logrus.Level(level), msg, args)
}

func (c *CustomLogger) Trace(msg string, args ...interface{}) {
	c.logger.Trace(msg, args)
}

func (c *CustomLogger) Debug(msg string, args ...interface{}) {
	c.logger.Debug(msg, args)
}

func (c *CustomLogger) Info(msg string, args ...interface{}) {
	c.logger.Info(args...)
}

func (c *CustomLogger) Warn(msg string, args ...interface{}) {
	c.logger.Warn(msg, args)
}

func (c *CustomLogger) Error(msg string, args ...interface{}) {
	c.logger.Error(msg, args)
}

func (c *CustomLogger) IsTrace() bool {
	return c.logger.Level == logrus.TraceLevel
}

func (c *CustomLogger) IsDebug() bool {
	return c.logger.Level == logrus.DebugLevel
}

func (c *CustomLogger) IsInfo() bool {
	return c.logger.Level == logrus.InfoLevel
}

func (c *CustomLogger) IsWarn() bool {
	return c.logger.Level == logrus.WarnLevel
}

func (c *CustomLogger) IsError() bool {
	return c.logger.Level == logrus.ErrorLevel
}

func (c *CustomLogger) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (c *CustomLogger) With(args ...interface{}) hclog.Logger {
	return c
}

func (c *CustomLogger) Name() string {
	return c.name
}

func (c *CustomLogger) Named(name string) hclog.Logger {
	return c
}

func (c *CustomLogger) ResetNamed(name string) hclog.Logger {
	return c
}

func (c *CustomLogger) SetLevel(level hclog.Level) {
	c.logger.SetLevel(logrus.Level(level))
}

func (c *CustomLogger) GetLevel() hclog.Level {
	return hclog.Level(c.logger.Level)
}

func (c *CustomLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(c.logger.Writer(), "", 0)
}

func (c *CustomLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return c.logger.Writer()
}
