package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/LXL47/jvmgo-playground/apps/internal/config"
	"github.com/LXL47/jvmgo-playground/apps/internal/protocol"
	"github.com/LXL47/jvmgo-playground/apps/internal/webui"
)

type Handler struct {
	cfg     config.API
	client  *http.Client
	limiter *limiter
	web     http.Handler
}

func NewHandler(cfg config.API) http.Handler {
	return securityHeaders(&Handler{
		cfg: cfg, client: &http.Client{Timeout: cfg.RunnerTimeout},
		limiter: newLimiter(cfg.RequestsPerMin), web: webui.Handler(),
	})
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/healthz" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case r.URL.Path == "/api/v1/runtime" && r.Method == http.MethodGet:
		h.proxy(w, r, http.MethodGet, "/v1/limits", nil)
	case r.URL.Path == "/api/v1/executions" && r.Method == http.MethodPost:
		h.execute(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/"):
		writeJSON(w, http.StatusNotFound, protocol.ErrorResponse{Error: "接口不存在"})
	default:
		h.web.ServeHTTP(w, r)
	}
}

func (h *Handler) execute(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		writeJSON(w, http.StatusUnsupportedMediaType, protocol.ErrorResponse{Error: "Content-Type 必须是 application/json"})
		return
	}
	if !h.limiter.Allow(clientIP(r), time.Now()) {
		w.Header().Set("Retry-After", "60")
		writeJSON(w, http.StatusTooManyRequests, protocol.ErrorResponse{Error: "请求过于频繁，请稍后重试"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxSourceBytes+1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request protocol.ExecuteRequest
	if err := decoder.Decode(&request); err != nil || int64(len(request.Source)) > h.cfg.MaxSourceBytes {
		writeJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "请求格式错误或源代码过大"})
		return
	}
	payload, _ := json.Marshal(request)
	h.proxy(w, r, http.MethodPost, "/v1/execute", payload)
}

func (h *Handler) proxy(w http.ResponseWriter, source *http.Request, method, path string, body []byte) {
	request, err := http.NewRequestWithContext(source.Context(), method, strings.TrimRight(h.cfg.RunnerURL, "/")+path, bytes.NewReader(body))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "无法创建内部请求"})
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := h.client.Do(request)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, protocol.ErrorResponse{Error: "执行服务暂不可用"})
		return
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, h.cfg.MaxSourceBytes+4096))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, protocol.ErrorResponse{Error: "读取执行服务响应失败"})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(response.StatusCode)
	_, _ = w.Write(data)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
