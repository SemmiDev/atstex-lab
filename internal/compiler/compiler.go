// Package compiler handles LaTeX source compilation to PDF.
// It supports pdflatex, xelatex, lualatex, and tectonic engines.
package compiler

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Engine represents a LaTeX compilation engine.
type Engine string

const (
	EnginePdfLatex Engine = "pdflatex"
	EngineXeLatex  Engine = "xelatex"
	EngineLuaLatex Engine = "lualatex"
	EngineTectonic Engine = "tectonic"

	defaultTimeout = 30 * time.Second
	maxSourceBytes = 512 * 1024 // 512 KB
)

// Result holds the output of a compilation run.
type Result struct {
	PDF     []byte
	Log     string
	Elapsed time.Duration
	Engine  Engine
}

// Options configures a compilation run.
type Options struct {
	Engine      Engine
	Timeout     time.Duration
	SinglePass  bool // Skip the second pass (faster preview, may have unresolved refs)
	PhotoBase64 string
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Engine:  EnginePdfLatex,
		Timeout: defaultTimeout,
	}
}

// Compile compiles LaTeX source and returns the resulting PDF bytes and log.
// It runs the engine twice to resolve references/labels (unless SinglePass is
// set), then reads the resulting PDF. Each invocation uses a fresh temporary
// directory that is cleaned up automatically.
func Compile(ctx context.Context, source []byte, opts Options) (*Result, error) {
	if len(source) == 0 {
		return nil, fmt.Errorf("source is empty")
	}
	if len(source) > maxSourceBytes {
		return nil, fmt.Errorf("source exceeds maximum allowed size (%d bytes)", maxSourceBytes)
	}

	engine := opts.Engine
	if engine == "" {
		engine = EnginePdfLatex
	}
	if _, err := exec.LookPath(string(engine)); err != nil {
		return nil, fmt.Errorf("engine %q not found in PATH: %w", engine, err)
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Work in a temp directory that is cleaned up automatically.
	workDir, err := os.MkdirTemp("", "atstex-lab-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	srcFile := filepath.Join(workDir, "document.tex")
	if err := os.WriteFile(srcFile, source, 0o600); err != nil {
		return nil, fmt.Errorf("writing source: %w", err)
	}

	// Handle optional base64 profile photo decoding.
	if opts.PhotoBase64 != "" {
		// A typical data URL looks like data:image/jpeg;base64,...
		b64data := opts.PhotoBase64
		ext := "jpg" // default
		if strings.HasPrefix(opts.PhotoBase64, "data:image/") {
			parts := strings.SplitN(opts.PhotoBase64, ",", 2)
			if len(parts) == 2 {
				// detect extension from "data:image/png;base64"
				if strings.Contains(parts[0], "png") {
					ext = "png"
				}
				b64data = parts[1]
			}
		}

		photoBytes, err := base64.StdEncoding.DecodeString(b64data)
		if err == nil {
			photoPath := filepath.Join(workDir, "photo."+ext)
			os.WriteFile(photoPath, photoBytes, 0o600)
		}
	}

	start := time.Now()

	// Tectonic uses a different invocation and handles multiple passes internally.
	if engine == EngineTectonic {
		return compileTectonic(ctx, workDir, srcFile, start, timeout)
	}

	// Determine number of passes: 2 for full compile, 1 for fast preview.
	passes := 2
	if opts.SinglePass {
		passes = 1
	}

	// Run the traditional TeX engine with security-hardened flags.
	var logBuf bytes.Buffer
	for pass := 1; pass <= passes; pass++ {
		logBuf.Reset()
		cmd := exec.CommandContext(ctx,
			string(engine),
			"-no-shell-escape",
			"-interaction=nonstopmode",
			"-halt-on-error",
			"-file-line-error",
			"-output-directory", workDir,
			srcFile,
		)
		cmd.Dir = workDir
		cmd.Stdout = &logBuf
		cmd.Stderr = &logBuf

		// Apply OS-level resource limits (Linux only).
		setProcessLimits(cmd)

		if runErr := cmd.Run(); runErr != nil {
			elapsed := time.Since(start)
			log := cleanLog(logBuf.String())
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("compilation timed out after %s", timeout)
			}
			return &Result{
				Log:     log,
				Elapsed: elapsed,
				Engine:  engine,
			}, fmt.Errorf("compilation failed (pass %d): %w", pass, runErr)
		}
	}

	elapsed := time.Since(start)

	pdfPath := filepath.Join(workDir, "document.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("reading output PDF: %w", err)
	}

	return &Result{
		PDF:     pdfBytes,
		Log:     cleanLog(logBuf.String()),
		Elapsed: elapsed,
		Engine:  engine,
	}, nil
}

// compileTectonic handles compilation via the Tectonic engine, which
// manages its own multi-pass logic internally.
func compileTectonic(ctx context.Context, workDir, srcFile string, start time.Time, timeout time.Duration) (*Result, error) {
	var logBuf bytes.Buffer

	cmd := exec.CommandContext(ctx,
		"tectonic",
		"--outdir", workDir,
		"--keep-logs",
		"--untrusted",
		srcFile,
	)
	cmd.Dir = workDir
	cmd.Stdout = &logBuf
	cmd.Stderr = &logBuf

	setProcessLimits(cmd)

	if runErr := cmd.Run(); runErr != nil {
		elapsed := time.Since(start)
		log := cleanLog(logBuf.String())
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("compilation timed out after %s", timeout)
		}
		return &Result{
			Log:     log,
			Elapsed: elapsed,
			Engine:  EngineTectonic,
		}, fmt.Errorf("tectonic compilation failed: %w", runErr)
	}

	elapsed := time.Since(start)

	pdfPath := filepath.Join(workDir, "document.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("reading output PDF: %w", err)
	}

	return &Result{
		PDF:     pdfBytes,
		Log:     cleanLog(logBuf.String()),
		Elapsed: elapsed,
		Engine:  EngineTectonic,
	}, nil
}

// cleanLog trims noise from the TeX log output.
func cleanLog(raw string) string {
	lines := strings.Split(raw, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Skip long path-dump lines that are just TeX searching for files.
		if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") && len(trimmed) > 80 {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
