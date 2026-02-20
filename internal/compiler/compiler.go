// Package compiler handles LaTeX source compilation to PDF.
// It supports pdflatex, xelatex, and lualatex engines.
package compiler

import (
	"bytes"
	"context"
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

	defaultTimeout = 60 * time.Second
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
	Engine  Engine
	Timeout time.Duration
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Engine:  EnginePdfLatex,
		Timeout: defaultTimeout,
	}
}

// Compile compiles LaTeX source and returns the resulting PDF bytes and log.
// It runs the engine twice to resolve references/labels, then once more if
// needed (BibTeX not included for simplicity — add as needed).
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
	workDir, err := os.MkdirTemp("", "latexpad-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	srcFile := filepath.Join(workDir, "document.tex")
	if err := os.WriteFile(srcFile, source, 0o600); err != nil {
		return nil, fmt.Errorf("writing source: %w", err)
	}

	start := time.Now()

	// Run twice so \ref, \label, \tableofcontents, etc. resolve correctly.
	var logBuf bytes.Buffer
	for pass := 1; pass <= 2; pass++ {
		logBuf.Reset()
		cmd := exec.CommandContext(ctx,
			string(engine),
			"-interaction=nonstopmode",
			"-halt-on-error",
			"-file-line-error",
			"-output-directory", workDir,
			srcFile,
		)
		cmd.Dir = workDir
		cmd.Stdout = &logBuf
		cmd.Stderr = &logBuf

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
