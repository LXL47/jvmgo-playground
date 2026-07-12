package runner

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/LXL47/jvmgo-playground/apps/internal/config"
	"github.com/LXL47/jvmgo-playground/apps/internal/protocol"
)

type Executor struct {
	cfg   config.Runner
	slots chan struct{}
}

type commandResult struct {
	output      string
	err         error
	timedOut    bool
	outputLimit bool
}

func NewExecutor(cfg config.Runner) *Executor {
	return &Executor{cfg: cfg, slots: make(chan struct{}, cfg.MaxConcurrent)}
}

func (e *Executor) Limits() protocol.Limits {
	return protocol.Limits{
		MaxInstructions: e.cfg.MaxInstructions,
		MaxHeapBytes:    e.cfg.MaxHeapBytes,
		MaxArrayLength:  e.cfg.MaxArrayLength,
		MaxOutputBytes:  e.cfg.MaxOutputBytes,
		TimeoutMS:       e.cfg.ExecutionTimeout.Milliseconds(),
	}
}

func (e *Executor) Execute(ctx context.Context, source string) (protocol.ExecuteResponse, int) {
	started := time.Now()
	response := protocol.ExecuteResponse{ID: newID(), Limits: e.Limits()}
	finish := func(status, output string, code int) (protocol.ExecuteResponse, int) {
		response.Status = status
		response.Output = output
		response.DurationMS = time.Since(started).Milliseconds()
		return response, code
	}

	if err := validateSource(source, e.cfg.MaxSourceBytes); err != nil {
		return finish("invalid_request", err.Error(), 400)
	}
	select {
	case e.slots <- struct{}{}:
		defer func() { <-e.slots }()
	default:
		return finish("busy", "执行队列已满，请稍后重试", 429)
	}

	workDir, err := os.MkdirTemp(e.cfg.WorkRoot, "job-"+response.ID+"-")
	if err != nil {
		return finish("internal_error", "无法创建任务目录", 500)
	}
	defer os.RemoveAll(workDir)
	if err := os.WriteFile(filepath.Join(workDir, "Main.java"), []byte(source), 0o600); err != nil {
		return finish("internal_error", "无法写入源代码", 500)
	}

	compile := runCommand(ctx, e.cfg.CompileTimeout, e.cfg.MaxOutputBytes, workDir, e.cfg.JavacPath,
		"-J-Xmx128m", "-J-Duser.language=en", "-J-Duser.country=US", "-J-Dfile.encoding=UTF-8",
		"-proc:none", "-encoding", "UTF-8", "-source", "8", "-target", "8",
		"-classpath", workDir, "-sourcepath", workDir, "-d", workDir, "Main.java")
	if compile.timedOut {
		return finish("compile_timeout", "编译超过时间限制", 408)
	}
	if compile.outputLimit {
		return finish("output_limit", compile.output, 413)
	}
	if compile.err != nil {
		return finish("compile_error", compile.output, 200)
	}
	if err := validateClasses(workDir, e.cfg.MaxClassFiles, e.cfg.MaxClassBytes); err != nil {
		return finish("compile_error", err.Error(), 200)
	}

	args := []string{
		"-sandbox",
		"-max-instructions", strconv.FormatUint(e.cfg.MaxInstructions, 10),
		"-max-heap-bytes", strconv.FormatUint(e.cfg.MaxHeapBytes, 10),
		"-max-array-length", strconv.FormatUint(e.cfg.MaxArrayLength, 10),
		"-max-output-bytes", strconv.FormatUint(e.cfg.MaxOutputBytes, 10),
		"-Xjre", e.cfg.JREPath, "-cp", workDir, "Main",
	}
	run := runCommand(ctx, e.cfg.ExecutionTimeout, e.cfg.MaxOutputBytes, workDir, e.cfg.JVMPath, args...)
	if run.timedOut {
		return finish("timeout", appendMessage(run.output, "执行超过时间限制"), 408)
	}
	if run.outputLimit {
		return finish("output_limit", appendMessage(run.output, "输出超过大小限制"), 413)
	}
	if run.err != nil {
		if strings.Contains(run.output, "sandbox:") {
			return finish("sandbox_limit", run.output, 200)
		}
		return finish("runtime_error", run.output, 200)
	}
	return finish("success", run.output, 200)
}

func validateSource(source string, maxBytes int64) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("Java 代码不能为空")
	}
	if int64(len(source)) > maxBytes {
		return fmt.Errorf("Java 代码不能超过 %d 字节", maxBytes)
	}
	if !utf8.ValidString(source) || strings.IndexByte(source, 0) >= 0 {
		return errors.New("Java 代码必须是有效 UTF-8 文本")
	}
	return nil
}

func validateClasses(root string, maxFiles int, maxBytes int64) error {
	var files int
	var bytes int64
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("编译产物不能包含符号链接")
		}
		if entry.IsDir() || filepath.Ext(path) != ".class" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		files++
		bytes += info.Size()
		if files > maxFiles || bytes > maxBytes {
			return errors.New("编译产物超过数量或大小限制")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if files == 0 {
		return errors.New("编译未生成 class 文件")
	}
	return nil
}

func runCommand(parent context.Context, timeout time.Duration, outputLimit uint64, dir, name string, args ...string) commandResult {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	output := newLimitedBuffer(outputLimit, cancel)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = output
	cmd.Stderr = output
	configureCommand(cmd)
	cmd.Cancel = func() error { return killCommand(cmd) }
	err := cmd.Run()
	return commandResult{
		output: strings.TrimSpace(output.String()), err: err,
		timedOut: errors.Is(ctx.Err(), context.DeadlineExceeded), outputLimit: output.Exceeded(),
	}
}

type limitedBuffer struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	limit    uint64
	exceeded bool
	cancel   context.CancelFunc
}

func newLimitedBuffer(limit uint64, cancel context.CancelFunc) *limitedBuffer {
	return &limitedBuffer{limit: limit, cancel: cancel}
}

func (b *limitedBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	original := len(data)
	remaining := int64(b.limit) - int64(b.buffer.Len())
	if remaining <= 0 {
		b.exceeded = true
		b.cancel()
		return original, nil
	}
	if int64(len(data)) > remaining {
		data = data[:remaining]
		b.exceeded = true
	}
	_, _ = b.buffer.Write(data)
	if b.exceeded {
		b.cancel()
	}
	return original, nil
}

func (b *limitedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

func (b *limitedBuffer) Exceeded() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.exceeded
}

func newID() string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(value[:])
}

func appendMessage(output, message string) string {
	if output == "" {
		return message
	}
	return output + "\n" + message
}
