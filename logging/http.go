package logging

import (
	"github.com/rs/zerolog/log"

	accesslog "github.com/mash/go-accesslog"
)

type HttpLogger struct {
}

func (l HttpLogger) Log(record accesslog.LogRecord) {
	log.Log().
		Str("method", record.Method).
		Str("request", record.Uri).
		Str("proto", record.Protocol).
		Str("remote_ip", record.Ip).
		Int("status", record.Status).
		Dur("request_time", record.ElapsedTime).
		Msg("")
}
