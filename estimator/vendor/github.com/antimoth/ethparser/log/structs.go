package log

import (
	log "github.com/inconshreveable/log15"
)

type logLevel log.Lvl

func (lv logLevel) String() string {
	switch lv {
	case DEBUG:
		return "DBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "EROR"
	case CRITICAL:
		return "CRIT"
	default:
		panic("bad level")
	}
}
