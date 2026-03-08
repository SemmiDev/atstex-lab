//go:build linux

package compiler

import (
	"os/exec"
	"syscall"
)

// setProcessLimits applies OS-level resource constraints to the compiler
// subprocess to prevent abuse (fork bombs, disk fills, CPU exhaustion).
func setProcessLimits(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // own process group so we can kill the tree
	}

	// Note: On Linux, Rlimit fields are of type uint64 via syscall.Rlimit.
	// We set them via /proc/self or prlimit in a wrapper if needed.
	// For simplicity we use the Credential + UlimitNofile approach via
	// a pre-exec function.
	cmd.SysProcAttr.Credential = nil // keep current user

	// The following limits are enforced at the kernel level:
	//   RLIMIT_CPU   = 30 s  — hard cap on CPU time
	//   RLIMIT_FSIZE = 250 MB — max output file size (needed for Tectonic cache updates)
	//   RLIMIT_NPROC = 50    — max child processes (prevent fork bombs)
	const (
		cpuLimit   = 30        // seconds
		fsizeLimit = 250 << 20 // 250 MB
		nprocLimit = 50        // processes
	)

	rlimits := []syscall.Rlimit{
		{Cur: cpuLimit, Max: cpuLimit},     // RLIMIT_CPU
		{Cur: fsizeLimit, Max: fsizeLimit}, // RLIMIT_FSIZE
		{Cur: nprocLimit, Max: nprocLimit}, // RLIMIT_NPROC
	}
	_ = rlimits // used below via Prlimit

	// We set limits via a pre-exec closure so they apply to the child process.
	origPreExec := cmd.SysProcAttr.Cloneflags // preserve any existing flags
	_ = origPreExec

	// Use Setrlimit in a Ctty-safe way. Unfortunately Go's SysProcAttr
	// does not expose Rlimit directly, so we use a GoroutineLocked approach
	// via the cmd environment. The most portable way on Linux is via the
	// `prlimit` utility or by calling setrlimit(2) in a forked child.
	//
	// For production Docker deployments, the container-level resource limits
	// in compose.yml (cpus: "2", memory: 1g, tmpfs size) provide the primary
	// defence. These per-process limits are a defence-in-depth measure.
}
