//go:build linux

package process

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/peterintech/nosleepp/internal/agent"
)

type linuxScanner struct {
	procRoot            string
	clockTicksPerSecond int
}

func NewScanner() Scanner {
	return linuxScanner{
		procRoot:            "/proc",
		clockTicksPerSecond: defaultLinuxClockTicksPerSecond,
	}
}

func (s linuxScanner) Scan(ctx context.Context) ([]agent.Process, error) {
	entries, err := os.ReadDir(s.procRoot)
	if err != nil {
		return nil, err
	}

	processes := make([]agent.Process, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !entry.IsDir() || !isNumeric(entry.Name()) {
			continue
		}

		proc, err := s.readProcess(entry.Name())
		if err != nil {
			if isProcRace(err) {
				continue
			}
			continue
		}
		processes = append(processes, proc)
	}

	return processes, nil
}

func (s linuxScanner) readProcess(pid string) (agent.Process, error) {
	statBytes, err := os.ReadFile(filepath.Join(s.procRoot, pid, "stat"))
	if err != nil {
		return agent.Process{}, err
	}

	proc, err := parseLinuxProcStat(string(statBytes), s.clockTicksPerSecond)
	if err != nil {
		return agent.Process{}, err
	}

	if commBytes, err := os.ReadFile(filepath.Join(s.procRoot, pid, "comm")); err == nil {
		if name := strings.TrimSpace(string(commBytes)); name != "" {
			proc.Name = name
		}
	}

	return proc, nil
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

func isProcRace(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission)
}
