package logger

import "log/slog"

func Error(err error) slog.Attr {
	return slog.Attr{Key: "err", Value: slog.AnyValue(err)}
}
