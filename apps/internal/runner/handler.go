package runner

import (
	"encoding/json"
	"net/http"

	"github.com/LXL47/jvmgo-playground/apps/internal/protocol"
)

type Handler struct {
	executor *Executor
	maxBody  int64
}

func NewHandler(executor *Executor, maxBody int64) http.Handler {
	return &Handler{executor: executor, maxBody: maxBody}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	if r.URL.Path == "/v1/limits" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, h.executor.Limits())
		return
	}
	if r.URL.Path == "/v1/execute" && r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBody+1024)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request protocol.ExecuteRequest
		if err := decoder.Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "请求必须是合法 JSON"})
			return
		}
		response, code := h.executor.Execute(r.Context(), request.Source)
		writeJSON(w, code, response)
		return
	}
	writeJSON(w, http.StatusNotFound, protocol.ErrorResponse{Error: "接口不存在"})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
