/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package logger

import (
	"log"
	"log/slog"
)

var (
	LogLevel                    = new(slog.LevelVar)
	DefaultHandler slog.Handler = NewHandler(&slog.HandlerOptions{
		AddSource: true,
		Level:     LogLevel,
	})
	DefaultSLogger     *slog.Logger = slog.New(DefaultHandler)
	DefaultLogger      *log.Logger  = slog.NewLogLogger(DefaultHandler, slog.LevelInfo)
	DefaultErrorLogger *log.Logger  = slog.NewLogLogger(DefaultHandler, slog.LevelError)
)
