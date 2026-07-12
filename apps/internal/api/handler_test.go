package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LXL47/jvmgo-playground/apps/internal/config"
)

func TestExecutionProxyAndSecurityHeaders(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/execute" {
			t.Fatalf("Runner 路径错误: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"job","status":"success","output":"ok","durationMs":1}`)
	}))
	defer runner.Close()

	handler := NewHandler(config.API{
		Listen: ":0", RunnerURL: runner.URL, RunnerTimeout: time.Second,
		MaxSourceBytes: 1024, RequestsPerMin: 10,
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/executions", strings.NewReader(`{"source":"public class Main {}"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"success"`) {
		t.Fatalf("代理响应错误: code=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Security-Policy") == "" || response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("缺少安全响应头")
	}
}

func TestLimiterRejectsAndCleansVisitors(t *testing.T) {
	limiter := newLimiter(1)
	limiter.maxVisitors = 1
	now := time.Now()
	if !limiter.Allow("first", now) || limiter.Allow("first", now) {
		t.Fatal("单地址限流错误")
	}
	if limiter.Allow("second", now) {
		t.Fatal("访问者上限应拒绝新地址")
	}
	if !limiter.Allow("second", now.Add(3*time.Minute)) {
		t.Fatal("过期访问者清理后应允许新地址")
	}
}
