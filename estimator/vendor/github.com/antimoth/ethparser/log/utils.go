package log

import (
	"fmt"
	"path"
	"runtime"
	"strings"
)

func GetRuntimeInfo(skip int) string {
	pc, fileP, line, ok := runtime.Caller(skip)
	if !ok {
		return "Get runtime caller error!"
	}

	callerStack := runtime.FuncForPC(pc).Name()
	var index int

	if index = strings.Index(callerStack, ".("); index == -1 {
		index = strings.LastIndex(callerStack, ".")
	}

	// parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	// pl := len(parts)

	// for i := 0; i < pl; i++ {
	// 	if parts[i][0] == '(' {
	// 		packageName = strings.Join(parts[0:i], ".")
	// 		break
	// 	}
	// }
	// if packageName == "" {
	//         packageName = strings.Join(parts[0:pl-1], ".")
	// }

	return fmt.Sprintf("%s@%s:%d", callerStack[0:index], path.Base(fileP), line)
}

func toLogLevel(lvl string) logLevel {
	switch lvl {
	case "debug", "DEBUG":
		return DEBUG
	case "info", "INFO":
		return INFO
	case "warn", "WARN":
		return WARN
	case "error", "ERROR":
		return ERROR
	case "critical", "CRITICAL":
		return CRITICAL
	default:
		return DEBUG
	}
}
