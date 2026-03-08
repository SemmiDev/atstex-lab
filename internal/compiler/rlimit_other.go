//go:build !linux

package compiler

import "os/exec"

// setProcessLimits is a no-op on non-Linux platforms (macOS, Windows).
// On these platforms, defence-in-depth is provided by the context timeout
// and Docker container resource limits (compose.yml).
func setProcessLimits(_ *exec.Cmd) {}
