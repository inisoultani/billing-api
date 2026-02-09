package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

func (h *Handler) ChangeLogLevel(level *slog.LevelVar) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		newLevel := r.URL.Query().Get("level")

		switch strings.ToUpper(newLevel) {
		case "DEBUG":
			level.Set(slog.LevelDebug)
		case "INFO":
			level.Set(slog.LevelInfo)
		case "WARN":
			level.Set(slog.LevelWarn)
		case "ERROR":
			level.Set(slog.LevelError)
		default:
			http.Error(w, "Invalid level", http.StatusBadRequest)
			return errors.New("")
		}

		slog.Debug("Log level changed", slog.String("new_level", strings.ToUpper(newLevel)))
		slog.Info("Log level changed", slog.String("new_level", strings.ToUpper(newLevel)))
		return json.NewEncoder(w).Encode(nil)
	}
}
