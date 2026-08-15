//go:build !windows && !darwin && !linux

package power

import "errors"

type unsupportedManager struct{}

func NewManager() Manager {
	return unsupportedManager{}
}

func (unsupportedManager) Acquire() error {
	return errors.New("power management is only supported on Windows, macOS, and Linux")
}

func (unsupportedManager) Release() error {
	return nil
}
