package log

import (
	log "github.com/inconshreveable/log15"
)

const (
	DEBUG    = logLevel(log.LvlDebug)
	INFO     = logLevel(log.LvlInfo)
	WARN     = logLevel(log.LvlWarn)
	ERROR    = logLevel(log.LvlError)
	CRITICAL = logLevel(log.LvlCrit)

	timeFormat  = "2006-01-02 15:04:05.000"
	termMsgJust = 40
)
