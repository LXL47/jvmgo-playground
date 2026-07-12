package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadRunner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runner.yaml")
	content := `runner:
  listen: ":8081"
  work_root: "/work"
  javac_path: "/jdk/bin/javac"
  jvm_path: "/app/jvmgo"
  jre_path: "/jdk/jre"
  compile_timeout: "5s"
  execution_timeout: "2s"
  max_source_bytes: 65536
  max_class_bytes: 2097152
  max_class_files: 64
  max_concurrent: 2
sandbox:
  max_instructions: 2000000
  max_heap_bytes: 33554432
  max_array_length: 100000
  max_output_bytes: 32768
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadRunner(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxInstructions != 2_000_000 || cfg.ExecutionTimeout.String() != "2s" {
		t.Fatalf("预算解析错误: %+v", cfg)
	}
}

func TestRepositoryConfigFiles(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", ".."))
	if _, err := LoadAPI(filepath.Join(root, "config", "api.yaml")); err != nil {
		t.Fatalf("默认 API 配置无效: %v", err)
	}
	if _, err := LoadRunner(filepath.Join(root, "config", "runner.yaml")); err != nil {
		t.Fatalf("默认 Runner 配置无效: %v", err)
	}
	for _, name := range []string{"compose.yaml", "compose.gvisor.yaml"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := yaml.Unmarshal(data, &document); err != nil {
			t.Fatalf("%s 不是合法 YAML: %v", name, err)
		}
		if document["services"] == nil {
			t.Fatalf("%s 缺少 services", name)
		}
	}
}

func TestLoadRunnerRejectsZeroBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runner.yaml")
	if err := os.WriteFile(path, []byte("runner:\n  compile_timeout: 1s\n  execution_timeout: 1s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRunner(path); err == nil {
		t.Fatal("空路径和零预算应被拒绝")
	}
}
