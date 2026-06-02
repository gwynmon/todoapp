package restapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// writeJSONError унифицирует возврат ошибок в формате JSON с логированием
func writeJSONError(w http.ResponseWriter, logger *slog.Logger, message string, code int) {
	if logger != nil {
		logger.Warn("api error", slog.String("message", message), slog.Int("status", code))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
