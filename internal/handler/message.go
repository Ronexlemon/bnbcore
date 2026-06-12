package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func parseDateParam(r *http.Request, key string, fallback time.Time) time.Time {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return fallback
	}
	return t
}