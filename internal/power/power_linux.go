//go:build linux

package power

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type linuxManager struct {
	mu  sync.Mutex
	cmd *exec.Cmd
}

func NewManager() Manager {
	return &linuxManager{}
}

func (m *linuxManager) Acquire() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != nil {
		return nil
	}

	cmd := exec.Command("systemd-inhibit", "--what=idle:sleep", "--why=nosleepp: agents are working", "--mode=block", "sleep", "infinity")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	m.cmd = cmd
	return nil
}

func (m *linuxManager) Release() error {
	m.mu.Lock()
	cmd := m.cmd
	m.cmd = nil
	m.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	killErr := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	waitErr := cmd.Wait()
	if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) && !errors.Is(killErr, syscall.ESRCH) {
		return killErr
	}
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			return nil
		}
		return waitErr
	}
	return nil
}
