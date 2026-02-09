package logger

import (
	bilingaApiHandler "billing-api/internal/http/handler"
	"log/slog"
	"os"
)

func NewLogger(isProduction bool) (*slog.Logger, *slog.LevelVar) {
	logLevel := &slog.LevelVar{}
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if isProduction {
		handler = slog.NewJSONHandler(os.Stdout, opts)
		logLevel.Set(slog.LevelInfo)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
		logLevel.Set(slog.LevelDebug)
	}

	wrapperHandler := &bilingaApiHandler.LogContextHandler{
		Handler: handler,
	}

	logger := slog.New(wrapperHandler)
	slog.SetDefault(logger)

	return logger, logLevel
}
