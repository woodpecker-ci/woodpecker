package errorattr

import "log/slog"

func Default(err error) slog.Attr {
	return slog.Attr{Key: "error", Value: slog.AnyValue(err)}
}
