package rpc_transport

import (
	"github.com/orbit-w/golib/bases/logger"
	"go.uber.org/zap"
)

/*
   @Author: orbit-w
   @File: logger
   @2024 1月 周二 23:40
*/

const (
	defaultFileName = "./logs/mm_rpc.log"
)

var (
	gLogger     = NewDefaultLogger()
	sugarLogger = gLogger.Sugar()
)

func Logger() *zap.Logger {
	return gLogger
}

func SugarLogger() *zap.SugaredLogger {
	return sugarLogger
}

func NewDefaultLogger() *zap.Logger {
	return logger.New(defaultFileName, zap.InfoLevel)
}
