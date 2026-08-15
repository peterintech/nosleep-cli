package process

import (
	"testing"
	"time"
)

func TestParseLinuxProcStat(t *testing.T) {
	proc, err := parseLinuxProcStat("123 (codex) S 45 1 1 0 -1 4194560 100 0 0 0 25 50 0 0 20 0 1 0 1000 0 0", 100)
	if err != nil {
		t.Fatalf("parseLinuxProcStat returned error: %v", err)
	}

	if proc.PID != 123 {
		t.Fatalf("expected PID 123, got %d", proc.PID)
	}
	if proc.ParentPID != 45 {
		t.Fatalf("expected parent PID 45, got %d", proc.ParentPID)
	}
	if proc.Name != "codex" {
		t.Fatalf("expected process name codex, got %q", proc.Name)
	}
	if proc.CPUTime != 750*time.Millisecond {
		t.Fatalf("expected 750ms CPU time, got %s", proc.CPUTime)
	}
}

func TestParseLinuxProcStatWithParenthesesInName(t *testing.T) {
	proc, err := parseLinuxProcStat("321 (agent (worker)) S 12 1 1 0 -1 0 0 0 0 0 10 10 0 0 20 0 1 0 1000 0 0", 100)
	if err != nil {
		t.Fatalf("parseLinuxProcStat returned error: %v", err)
	}

	if proc.Name != "agent (worker)" {
		t.Fatalf("expected process name with parentheses, got %q", proc.Name)
	}
}

func TestParseLinuxProcStatRejectsMalformedInput(t *testing.T) {
	if _, err := parseLinuxProcStat("not a proc stat line", 100); err == nil {
		t.Fatal("expected malformed input to fail")
	}
}

func TestTicksToDuration(t *testing.T) {
	got := ticksToDuration(250, 100)
	if got != 2500*time.Millisecond {
		t.Fatalf("expected 2.5s, got %s", got)
	}
}
