package log

import (
	"os"

	"bytes"
	"fmt"

	log "github.com/inconshreveable/log15"
)

func printLogCtx(buf *bytes.Buffer, ctx []interface{}) {
	for i := 0; i < len(ctx); i += 2 {
		k, ok := ctx[i].(string)
		v := fmt.Sprintf("%+v", ctx[i+1])
		if !ok {
			k, v = "ERR", fmt.Sprintf("%+v", k)
		}
		fmt.Fprintf(buf, " %s=%s", k, v)
	}
	buf.WriteByte('\n')
}

func terminalFormat() log.Format {
	return log.FormatFunc(func(r *log.Record) []byte {
		b := &bytes.Buffer{}
		level := logLevel(r.Lvl)
		fmt.Fprintf(b, "[%v][%s] %s", level, r.Time.Format(timeFormat), r.Msg)
		r.Ctx = append(r.Ctx, "caller", GetRuntimeInfo(14))
		if len(r.Ctx) > 0 && len(r.Msg) < termMsgJust {
			b.Write(bytes.Repeat([]byte{' '}, termMsgJust-len(r.Msg)))
		}
		printLogCtx(b, r.Ctx)
		return b.Bytes()
	})
}

// New 返回一个logger对象
func NewLogger(lvl string) log.Logger {
	handler := log.LvlFilterHandler(
		log.Lvl(toLogLevel(lvl)),
		log.StreamHandler(os.Stdout, terminalFormat()),
	)
	logger := log.New()
	logger.SetHandler(handler)
	return logger
}
