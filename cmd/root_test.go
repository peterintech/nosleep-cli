package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/peterintech/nosleepp/internal/agent"
)

type commandFakeScanner struct {
	calls int
	sets  [][]agent.Process
}

func (s *commandFakeScanner) Scan(ctx context.Context) ([]agent.Process, error) {
	if s.calls >= len(s.sets) {
		return s.sets[len(s.sets)-1], nil
	}
	next := s.sets[s.calls]
	s.calls++
	return next, nil
}

type commandFakePower struct {
	acquires int
	releases int
}

func (p *commandFakePower) Acquire() error {
	p.acquires++
	return nil
}

func (p *commandFakePower) Release() error {
	p.releases++
	return nil
}

func TestParseAgentFlag(t *testing.T) {
	profile, err := parseAgentFlag("Codex=codex,Codex.exe")
	if err != nil {
		t.Fatalf("parseAgentFlag returned error: %v", err)
	}
	if profile.Name != "Codex" || len(profile.Processes) != 2 {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestListCommandJSON(t *testing.T) {
	var out bytes.Buffer
	opts := &options{interval: time.Millisecond, sample: 0, output: &out, errorOutput: &bytes.Buffer{}}
	opts.processScan = &commandFakeScanner{sets: [][]agent.Process{
		{{PID: 7, Name: "codex.exe", CPUTime: 100 * time.Millisecond}},
		{{PID: 7, Name: "codex.exe", CPUTime: 500 * time.Millisecond}},
	}}
	root := newRootCommand(opts)
	root.SetArgs([]string{"list", "--json", "--sample", "0s"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(out.String(), `"agent": "Codex"`) {
		t.Fatalf("expected JSON output to include Codex, got %s", out.String())
	}
}

func TestWatchOnceNoAgentsExitCode(t *testing.T) {
	var out bytes.Buffer
	opts := &options{interval: time.Millisecond, sample: 0, output: &out, errorOutput: &bytes.Buffer{}}
	opts.processScan = &commandFakeScanner{sets: [][]agent.Process{{}, {}}}
	opts.powerManager = &commandFakePower{}
	root := newRootCommand(opts)
	root.SetArgs([]string{"watch", "--once", "--sample", "0s"})

	err := root.Execute()
	var exitErr ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if exitErr.Code != 1 {
		t.Fatalf("expected exit code 1, got %d", exitErr.Code)
	}
}

func TestPowerTestAcquiresAndReleases(t *testing.T) {
	var out bytes.Buffer
	power := &commandFakePower{}
	opts := &options{output: &out, errorOutput: &bytes.Buffer{}}
	opts.powerManager = power
	root := newRootCommand(opts)
	root.SetArgs([]string{"power-test", "--duration", "10ms"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if power.acquires != 1 {
		t.Fatalf("expected one acquire, got %d", power.acquires)
	}
	if power.releases != 1 {
		t.Fatalf("expected one release, got %d", power.releases)
	}
	if !strings.Contains(out.String(), "No-sleep lock acquired for 10ms") {
		t.Fatalf("expected lock-acquired message, got %s", out.String())
	}
	if !strings.Contains(out.String(), "No-sleep lock released") {
		t.Fatalf("expected lock-released message, got %s", out.String())
	}
}

func TestPowerTestRejectsNonPositiveDuration(t *testing.T) {
	for _, duration := range []string{"0s", "-5s"} {
		var out bytes.Buffer
		power := &commandFakePower{}
		opts := &options{output: &out, errorOutput: &bytes.Buffer{}}
		opts.powerManager = power
		root := newRootCommand(opts)
		root.SetArgs([]string{"power-test", "--duration=" + duration})

		err := root.Execute()
		var exitErr ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("expected ExitError for duration %q, got %v", duration, err)
		}
		if exitErr.Code != 2 {
			t.Fatalf("expected exit code 2 for duration %q, got %d", duration, exitErr.Code)
		}
		if power.acquires != 0 || power.releases != 0 {
			t.Fatalf("duration %q should not touch power state, got acquire=%d release=%d", duration, power.acquires, power.releases)
		}
	}
}
