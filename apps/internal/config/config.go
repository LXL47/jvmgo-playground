package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type API struct {
	Listen         string
	RunnerURL      string
	RunnerTimeout  time.Duration
	MaxSourceBytes int64
	RequestsPerMin int
}

type Runner struct {
	Listen           string
	WorkRoot         string
	JavacPath        string
	JVMPath          string
	JREPath          string
	CompileTimeout   time.Duration
	ExecutionTimeout time.Duration
	MaxSourceBytes   int64
	MaxClassBytes    int64
	MaxClassFiles    int
	MaxConcurrent    int
	MaxInstructions  uint64
	MaxHeapBytes     uint64
	MaxArrayLength   uint64
	MaxOutputBytes   uint64
}

type apiFile struct {
	API struct {
		Listen         string `yaml:"listen"`
		RunnerURL      string `yaml:"runner_url"`
		RunnerTimeout  string `yaml:"runner_timeout"`
		MaxSourceBytes int64  `yaml:"max_source_bytes"`
		RequestsPerMin int    `yaml:"requests_per_minute"`
	} `yaml:"api"`
}

type runnerFile struct {
	Runner struct {
		Listen           string `yaml:"listen"`
		WorkRoot         string `yaml:"work_root"`
		JavacPath        string `yaml:"javac_path"`
		JVMPath          string `yaml:"jvm_path"`
		JREPath          string `yaml:"jre_path"`
		CompileTimeout   string `yaml:"compile_timeout"`
		ExecutionTimeout string `yaml:"execution_timeout"`
		MaxSourceBytes   int64  `yaml:"max_source_bytes"`
		MaxClassBytes    int64  `yaml:"max_class_bytes"`
		MaxClassFiles    int    `yaml:"max_class_files"`
		MaxConcurrent    int    `yaml:"max_concurrent"`
	} `yaml:"runner"`
	Sandbox struct {
		MaxInstructions uint64 `yaml:"max_instructions"`
		MaxHeapBytes    uint64 `yaml:"max_heap_bytes"`
		MaxArrayLength  uint64 `yaml:"max_array_length"`
		MaxOutputBytes  uint64 `yaml:"max_output_bytes"`
	} `yaml:"sandbox"`
}

func LoadAPI(path string) (API, error) {
	var raw apiFile
	if err := load(path, &raw); err != nil {
		return API{}, err
	}
	timeout, err := positiveDuration("api.runner_timeout", raw.API.RunnerTimeout)
	if err != nil {
		return API{}, err
	}
	cfg := API{
		Listen: raw.API.Listen, RunnerURL: raw.API.RunnerURL, RunnerTimeout: timeout,
		MaxSourceBytes: raw.API.MaxSourceBytes, RequestsPerMin: raw.API.RequestsPerMin,
	}
	if cfg.Listen == "" || cfg.RunnerURL == "" || cfg.MaxSourceBytes <= 0 || cfg.RequestsPerMin <= 0 {
		return API{}, errors.New("api 配置包含空值或非正数")
	}
	return cfg, nil
}

func LoadRunner(path string) (Runner, error) {
	var raw runnerFile
	if err := load(path, &raw); err != nil {
		return Runner{}, err
	}
	compileTimeout, err := positiveDuration("runner.compile_timeout", raw.Runner.CompileTimeout)
	if err != nil {
		return Runner{}, err
	}
	executionTimeout, err := positiveDuration("runner.execution_timeout", raw.Runner.ExecutionTimeout)
	if err != nil {
		return Runner{}, err
	}
	cfg := Runner{
		Listen: raw.Runner.Listen, WorkRoot: raw.Runner.WorkRoot, JavacPath: raw.Runner.JavacPath,
		JVMPath: raw.Runner.JVMPath, JREPath: raw.Runner.JREPath, CompileTimeout: compileTimeout,
		ExecutionTimeout: executionTimeout, MaxSourceBytes: raw.Runner.MaxSourceBytes,
		MaxClassBytes: raw.Runner.MaxClassBytes, MaxClassFiles: raw.Runner.MaxClassFiles,
		MaxConcurrent: raw.Runner.MaxConcurrent, MaxInstructions: raw.Sandbox.MaxInstructions,
		MaxHeapBytes: raw.Sandbox.MaxHeapBytes, MaxArrayLength: raw.Sandbox.MaxArrayLength,
		MaxOutputBytes: raw.Sandbox.MaxOutputBytes,
	}
	base := filepath.Dir(path)
	cfg.WorkRoot = resolvePath(base, cfg.WorkRoot)
	cfg.JavacPath = resolvePath(base, cfg.JavacPath)
	cfg.JVMPath = resolvePath(base, cfg.JVMPath)
	cfg.JREPath = resolvePath(base, cfg.JREPath)
	if cfg.Listen == "" || cfg.WorkRoot == "" || cfg.JavacPath == "" || cfg.JVMPath == "" || cfg.JREPath == "" {
		return Runner{}, errors.New("runner 路径配置不能为空")
	}
	if cfg.MaxSourceBytes <= 0 || cfg.MaxClassBytes <= 0 || cfg.MaxClassFiles <= 0 || cfg.MaxConcurrent <= 0 ||
		cfg.MaxInstructions == 0 || cfg.MaxHeapBytes == 0 || cfg.MaxArrayLength == 0 || cfg.MaxOutputBytes == 0 {
		return Runner{}, errors.New("runner 预算必须为正数")
	}
	return cfg, nil
}

func resolvePath(base, value string) string {
	if value == "" || filepath.IsAbs(value) {
		return value
	}
	resolved, err := filepath.Abs(filepath.Join(base, value))
	if err != nil {
		return filepath.Clean(filepath.Join(base, value))
	}
	return resolved
}

func load(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置 %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("解析配置 %s: %w", path, err)
	}
	return nil
}

func positiveDuration(name, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s 必须是正数时长", name)
	}
	return duration, nil
}
