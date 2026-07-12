package runner

import (
	"context"
	"strings"
	"testing"
)

func TestValidateSource(t *testing.T) {
	if err := validateSource("public class Main {}", 64); err != nil {
		t.Fatal(err)
	}
	if err := validateSource("   ", 64); err == nil {
		t.Fatal("空源码应被拒绝")
	}
	if err := validateSource(strings.Repeat("a", 65), 64); err == nil {
		t.Fatal("超长源码应被拒绝")
	}
}

func TestLimitedBufferCancelsAtLimit(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	buffer := newLimitedBuffer(4, cancel)
	if _, err := buffer.Write([]byte("abcdef")); err != nil {
		t.Fatal(err)
	}
	if !buffer.Exceeded() || buffer.String() != "abcd" {
		t.Fatalf("输出限制错误: exceeded=%v output=%q", buffer.Exceeded(), buffer.String())
	}
}
