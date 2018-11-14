package log

import (
	log "github.com/inconshreveable/log15"
)

var logger = log.New()

func LogError(sDescribe string, args ...interface{}) {
	logger.Error(sDescribe, append(args, "caller", GetRuntimeInfo(2))...)
}

func LogInfo(sDescribe string, args ...interface{}) {
	logger.Info(sDescribe, append(args, "caller", GetRuntimeInfo(2))...)
}

func LogDebug(sDescribe string, args ...interface{}) {
	logger.Debug(sDescribe, append(args, "caller", GetRuntimeInfo(2))...)
}

func LogWarn(sDescribe string, args ...interface{}) {
	logger.Warn(sDescribe, append(args, "caller", GetRuntimeInfo(2))...)
}
