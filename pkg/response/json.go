package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Meta  any    `json:"meta,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(Envelope{Data: data}); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

func Error(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(Envelope{Error: msg}); err != nil {
		slog.Error("failed to encode error response", "error", err)
	}
}

func JSONWithMeta(w http.ResponseWriter, status int, data, meta any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(Envelope{Data: data, Meta: meta}); err != nil {
		slog.Error("failed to encode JSON response with meta", "error", err)
	}
}

func DecodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}
