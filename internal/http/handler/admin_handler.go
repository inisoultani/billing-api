package handler

import (
	"encoding/json"
	"errors"
	"fmt"
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
			return BadRequest(fmt.Sprintf("Invalid level : %s", newLevel), errors.New("Invalid level"))
		}

		slog.Debug("Log level changed", slog.String("new_level", strings.ToUpper(newLevel)))
		slog.Info("Log level changed", slog.String("new_level", strings.ToUpper(newLevel)))

		return json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Log level changed to %s", level),
		})
	}
}
