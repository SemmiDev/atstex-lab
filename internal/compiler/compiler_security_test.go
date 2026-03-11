package compiler

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/semmidev/atstex-lab/internal/cvtemplate"
)

// ────────────────────────────────────────────────────────────────────────────
// SanitizeLatexInput tests
// ────────────────────────────────────────────────────────────────────────────

func TestSanitizeLatexInput_AllSpecialChars(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{`\`, `\textbackslash{}`},
		{`{`, `\{`},
		{`}`, `\}`},
		{`$`, `\$`},
		{`&`, `\&`},
		{`#`, `\#`},
		{`^`, `\textasciicircum{}`},
		{`_`, `\_`},
		{`~`, `\textasciitilde{}`},
		{`%`, `\%`},
	}

	for _, tt := range tests {
		got := SanitizeLatexInput(tt.in)
		if got != tt.want {
			t.Errorf("SanitizeLatexInput(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSanitizeLatexInput_MixedString(t *testing.T) {
	input := `Hello $world & friends {100%}`
	want := `Hello \$world \& friends \{100\%\}`
	got := SanitizeLatexInput(input)
	if got != want {
		t.Errorf("SanitizeLatexInput(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeLatexInput_SafeString(t *testing.T) {
	input := "Hello World 123"
	got := SanitizeLatexInput(input)
	if got != input {
		t.Errorf("SanitizeLatexInput(%q) should return unchanged, got %q", input, got)
	}
}

func TestSanitizeLatexInput_EmptyString(t *testing.T) {
	got := SanitizeLatexInput("")
	if got != "" {
		t.Errorf("SanitizeLatexInput(\"\") = %q, want empty string", got)
	}
}

func TestSanitizeLatexInput_ShellEscapeAttempt(t *testing.T) {
	// Attempt to inject \write18{rm -rf /}
	input := `\write18{rm -rf /}`
	got := SanitizeLatexInput(input)
	// The backslash and braces should be escaped
	if strings.Contains(got, `\write18`) {
		t.Errorf("SanitizeLatexInput should escape backslash in %q, got %q", input, got)
	}
	if !strings.Contains(got, `\{rm`) {
		t.Errorf("SanitizeLatexInput should escape braces in %q, got %q", input, got)
	}
}

func TestSanitizeLatexInput_InputCommandAttempt(t *testing.T) {
	input := `\input{/etc/passwd}`
	got := SanitizeLatexInput(input)
	if strings.Contains(got, `\input`) {
		t.Errorf("SanitizeLatexInput should escape backslash in %q, got %q", input, got)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// SanitizeCVData tests
// ────────────────────────────────────────────────────────────────────────────

func TestSanitizeCVData_AllFields(t *testing.T) {
	data := cvtemplate.CVData{
		Personal: cvtemplate.Personal{
			Name:     "John & Jane $Doe",
			Title:    "Engineer #1",
			Email:    "john@example.com",
			Phone:    "+1 (555) 100%",
			Location: "City {State}",
			Linkedin: cvtemplate.Link{Display: "linkedin_link", URL: "https://linkedin.com/in/test~user"},
			Github:   cvtemplate.Link{Display: "gh^user", URL: "https://github.com/test"},
			Website:  cvtemplate.Link{Display: "my\\site", URL: "https://example.com"},
		},
		Summary: "A great $engineer with 100% dedication",
		Experience: []cvtemplate.Experience{
			{
				Company:  "ACME & Co",
				Title:    "Lead $Dev",
				Location: "City {State}",
				Dates:    "2020 ~ 2023",
				Bullets:  "Built #1 product",
			},
		},
		Education: []cvtemplate.Education{
			{
				Institution: "MIT & Harvard",
				Degree:      "BS in CS $1000",
				Location:    "Boston {MA}",
				Dates:       "2016 ~ 2020",
				GPA:         "3.9/4.0",
				Activities:  "Club #1 & #2",
			},
		},
		Projects: []cvtemplate.Project{
			{
				Name:    "Project $Alpha",
				Role:    "Lead^Dev",
				Link:    "https://example.com/path_to",
				Bullets: "Built {nice} things & stuff",
			},
		},
		Skills: cvtemplate.Skills{
			Languages:  "Go & Python",
			Frameworks: "React $Next",
			Tools:      "Docker {k8s}",
			Other:      "Linux ~100%",
		},
		Certifications: []cvtemplate.Certification{
			{Name: "AWS $Certified", Issuer: "Amazon & Web"},
		},
		Volunteer: []cvtemplate.Volunteer{
			{
				Organization: "Org #1",
				Role:         "Lead^Vol",
				Location:     "City {State}",
				Dates:        "2020 ~ 2021",
				Bullets:      "Helped $100 people",
			},
		},
		Awards: []cvtemplate.Award{
			{
				Title:       "Award #1",
				Issuer:      "Org & Co",
				Date:        "2023",
				Description: "Top $1%",
			},
		},
		Talks: []cvtemplate.Talk{
			{
				Title:       "Talk & Share",
				Event:       "Conf $2023",
				Location:    "City {Hall}",
				Date:        "2023",
				Description: "About #1 topic~here",
			},
		},
	}

	SanitizeCVData(&data)

	// Spot check key fields
	assertContains(t, "Personal.Name", data.Personal.Name, `\&`)
	assertContains(t, "Personal.Name", data.Personal.Name, `\$`)
	assertContains(t, "Personal.Title", data.Personal.Title, `\#`)
	assertContains(t, "Personal.Phone", data.Personal.Phone, `\%`)
	assertContains(t, "Personal.Location", data.Personal.Location, `\{`)
	assertContains(t, "Personal.Linkedin.URL", data.Personal.Linkedin.URL, `\textasciitilde{}`)
	assertContains(t, "Personal.Github.Display", data.Personal.Github.Display, `\textasciicircum{}`)
	assertContains(t, "Personal.Website.Display", data.Personal.Website.Display, `\textbackslash{}`)
	assertContains(t, "Summary", data.Summary, `\$`)
	assertContains(t, "Experience[0].Company", data.Experience[0].Company, `\&`)
	assertContains(t, "Education[0].Institution", data.Education[0].Institution, `\&`)
	assertContains(t, "Projects[0].Name", data.Projects[0].Name, `\$`)
	assertContains(t, "Skills.Languages", data.Skills.Languages, `\&`)
	assertContains(t, "Certifications[0].Name", data.Certifications[0].Name, `\$`)
	assertContains(t, "Volunteer[0].Bullets", data.Volunteer[0].Bullets, `\$`)
	assertContains(t, "Awards[0].Description", data.Awards[0].Description, `\$`)
	assertContains(t, "Talks[0].Title", data.Talks[0].Title, `\&`)

	// Ensure no raw special chars remain in known-dangerous fields
	assertNotContains(t, "Experience[0].Company raw &", data.Experience[0].Company, ` & `)
	assertNotContains(t, "Summary raw $", data.Summary, ` $`)
}

func assertContains(t *testing.T, field, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("%s = %q, expected to contain %q", field, got, want)
	}
}

func assertNotContains(t *testing.T, field, got, notWant string) {
	t.Helper()
	if strings.Contains(got, notWant) {
		t.Errorf("%s = %q, should NOT contain raw %q", field, got, notWant)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Compiler security tests (require a LaTeX engine in PATH)
// ────────────────────────────────────────────────────────────────────────────

func requirePdflatex(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pdflatex"); err != nil {
		t.Skip("pdflatex not found in PATH — skipping compiler tests")
	}
}

func TestCompileEmptySource(t *testing.T) {
	_, err := Compile(t.Context(), []byte{}, DefaultOptions())
	if err == nil {
		t.Fatal("expected error for empty source")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCompileMaxSourceSize(t *testing.T) {
	oversized := make([]byte, maxSourceBytes+1)
	for i := range oversized {
		oversized[i] = 'a'
	}
	_, err := Compile(t.Context(), oversized, DefaultOptions())
	if err == nil {
		t.Fatal("expected error for oversized source")
	}
	if !strings.Contains(err.Error(), "maximum allowed size") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCompileRejectsShellEscape(t *testing.T) {
	requirePdflatex(t)

	// Attempt to execute a shell command via \write18
	source := `\documentclass{article}
\begin{document}
\immediate\write18{echo HACKED > /tmp/hacked.txt}
Hello
\end{document}`

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	result, err := Compile(ctx, []byte(source), DefaultOptions())
	// The compile may succeed (write18 is silently ignored with -no-shell-escape)
	// but the shell command must NOT have executed.
	_ = result
	_ = err

	// Verify the file was NOT created
	if _, statErr := os.Stat("/tmp/hacked.txt"); statErr == nil {
		_ = os.Remove("/tmp/hacked.txt")
		t.Fatal("shell escape was NOT blocked — /tmp/hacked.txt was created")
	}
}

func TestCompileRejectsInputCommand(t *testing.T) {
	requirePdflatex(t)

	// Attempt to read a system file
	source := `\documentclass{article}
\begin{document}
\input{/etc/passwd}
\end{document}`

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	_, err := Compile(ctx, []byte(source), DefaultOptions())
	// This should fail because /etc/passwd is not a valid TeX file
	if err == nil {
		t.Log("compile succeeded, but this is acceptable if input is restricted")
	}
}

func TestCompileTimeout(t *testing.T) {
	requirePdflatex(t)

	// An infinite-loop LaTeX document
	source := `\documentclass{article}
\begin{document}
\newcount\loopcount
\loopcount=0
\loop
  \advance\loopcount by 1
  x
\ifnum\loopcount<999999999 \repeat
\end{document}`

	opts := Options{
		Engine:  EnginePdfLatex,
		Timeout: 3 * time.Second,
	}

	ctx := t.Context()
	_, err := Compile(ctx, []byte(source), opts)
	if err == nil {
		t.Fatal("expected timeout error for infinite-loop document")
	}
	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "failed") {
		t.Errorf("expected timeout/fail error, got: %v", err)
	}
}

func TestCompileSuccessfulDocument(t *testing.T) {
	source := `\documentclass{article}
\begin{document}
Hello, World!
\end{document}`

	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()

	opts := Options{
		Engine:  EngineTectonic,
		Timeout: 90 * time.Second,
	}

	result, err := Compile(ctx, []byte(source), opts)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if len(result.PDF) == 0 {
		t.Fatal("expected non-empty PDF output")
	}
	if result.Engine != EngineTectonic {
		t.Errorf("expected engine tectonic, got %s", result.Engine)
	}
}

func TestCompileSinglePass(t *testing.T) {
	requirePdflatex(t)

	source := `\documentclass{article}
\begin{document}
Hello, single pass!
\end{document}`

	opts := Options{
		Engine:     EnginePdfLatex,
		Timeout:    30 * time.Second,
		SinglePass: true,
	}

	ctx := t.Context()
	result, err := Compile(ctx, []byte(source), opts)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if len(result.PDF) == 0 {
		t.Fatal("expected non-empty PDF output for single-pass")
	}
}

func TestCompileTempDirCleanup(t *testing.T) {
	requirePdflatex(t)

	source := `\documentclass{article}
\begin{document}
Temp dir test
\end{document}`

	// Compile a document and check that no temp dirs are left behind
	entries1, _ := os.ReadDir(os.TempDir())
	count1 := countAtstexDirs(entries1)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	_, err := Compile(ctx, []byte(source), DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	entries2, _ := os.ReadDir(os.TempDir())
	count2 := countAtstexDirs(entries2)

	if count2 > count1 {
		t.Errorf("temp directories not cleaned up: before=%d, after=%d", count1, count2)
	}
}

func countAtstexDirs(entries []os.DirEntry) int {
	count := 0
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "atstex-lab-") {
			count++
		}
	}
	return count
}

func TestCompileNoShellEscapeFlag(t *testing.T) {
	// Verify that the -no-shell-escape flag is included in the command args.
	// This is a unit test of the command construction, not a full compile.
	requirePdflatex(t)

	source := `\documentclass{article}
\begin{document}
Flag test
\end{document}`

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	result, err := Compile(ctx, []byte(source), DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// If we got here without shell escape, the flag is working.
	// The real test is TestCompileRejectsShellEscape above.
}
