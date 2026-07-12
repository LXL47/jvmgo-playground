//go:build windows

package runner

import "os/exec"

func configureCommand(_ *exec.Cmd) {}

func killCommand(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
