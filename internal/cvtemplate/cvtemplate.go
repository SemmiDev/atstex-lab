package cvtemplate

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/semmidev/atstex-lab/web"
)

// TemplateInfo holds metadata about a CV template.
type TemplateInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// List returns all available CV templates.
func List() ([]TemplateInfo, error) {
	entries, err := fs.ReadDir(web.TemplateFS, "templates/cv")
	if err != nil {
		// If dir doesn't exist, just return empty list
		return []TemplateInfo{}, nil
	}

	var tpls []TemplateInfo
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".tex") {
			name := strings.TrimSuffix(e.Name(), ".tex")
			tpls = append(tpls, TemplateInfo{
				Name:        name,
				Description: strings.ToTitle(name),
			})
		}
	}
	return tpls, nil
}

// Get returns the source code of a specific template.
func Get(name string) (string, error) {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, ".") {
		return "", fmt.Errorf("invalid template name")
	}
	path := filepath.Join("templates/cv", name+".tex")
	b, err := fs.ReadFile(web.TemplateFS, path)
	if err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}
	return string(b), nil
}

type Personal struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
	Linkedin string `json:"linkedin"`
	Github   string `json:"github"`
	Website  string `json:"website"`
}

type Experience struct {
	Company  string `json:"company"`
	Title    string `json:"title"`
	Location string `json:"location"`
	Dates    string `json:"dates"`
	Bullets  string `json:"bullets"`
}

type Education struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Dates       string `json:"dates"`
	GPA         string `json:"gpa"`
}

type Project struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Link    string `json:"link"`
	Bullets string `json:"bullets"`
}

type Skills struct {
	Languages  string `json:"languages"`
	Frameworks string `json:"frameworks"`
	Tools      string `json:"tools"`
	Other      string `json:"other"`
}

type Certification struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
}

type CVData struct {
	Personal       Personal        `json:"personal"`
	Summary        string          `json:"summary"`
	Experience     []Experience    `json:"experience"`
	Education      []Education     `json:"education"`
	Projects       []Project       `json:"projects"`
	Skills         Skills          `json:"skills"`
	Certifications []Certification `json:"certifications"`
}

var texEscapes = strings.NewReplacer(
	"\\", "\\textbackslash{}",
	"&", "\\&",
	"%", "\\%",
	"$", "\\$",
	"#", "\\#",
	"_", "\\_",
	"{", "\\{",
	"}", "\\}",
	"~", "\\textasciitilde{}",
	"^", "\\textasciicircum{}",
)

// texEscape escapes special characters for LaTeX.
func texEscape(s string) string {
	return texEscapes.Replace(s)
}

// texLines splits a string by newline and trims space.
func texLines(s string) []string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// Render processes the given LaTeX template with the provided CVData.
func Render(name string, data CVData) (string, error) {
	raw, err := Get(name)
	if err != nil {
		return "", err
	}

	t := template.New(name).Delims("[[", "]]").Funcs(template.FuncMap{
		"texEscape": texEscape,
		"texLines":  texLines,
	})

	t, err = t.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
