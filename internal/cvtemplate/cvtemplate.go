package cvtemplate

import (
	"bytes"
	"errors"
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
		//nolint:nilerr // If dir doesn't exist, just return empty list
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
		return "", errors.New("invalid template name")
	}
	path := filepath.Join("templates/cv", name+".tex")
	b, err := fs.ReadFile(web.TemplateFS, path)
	if err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}
	return string(b), nil
}

type Link struct {
	Display string `json:"display"`
	URL     string `json:"url"`
}

type Personal struct {
	Photo    string `json:"photo"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
	Linkedin Link   `json:"linkedin"`
	Github   Link   `json:"github"`
	Website  Link   `json:"website"`
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
	Location    string `json:"location"`
	Dates       string `json:"dates"`
	GPA         string `json:"gpa"`
	Activities  string `json:"activities"`
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

type Volunteer struct {
	Organization string `json:"organization"`
	Role         string `json:"role"`
	Location     string `json:"location"`
	Dates        string `json:"dates"`
	Bullets      string `json:"bullets"`
}

type Award struct {
	Title       string `json:"title"`
	Issuer      string `json:"issuer"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

type Talk struct {
	Title       string `json:"title"`
	Event       string `json:"event"`
	Location    string `json:"location"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

type CVData struct {
	Personal       Personal        `json:"personal"`
	Summary        string          `json:"summary"`
	Experience     []Experience    `json:"experience"`
	Education      []Education     `json:"education"`
	Projects       []Project       `json:"projects"`
	Skills         Skills          `json:"skills"`
	Certifications []Certification `json:"certifications"`
	Volunteer      []Volunteer     `json:"volunteer"`
	Awards         []Award         `json:"awards"`
	Talks          []Talk          `json:"talks"`
}

// PageSettings holds user-customisable document layout overrides.
type PageSettings struct {
	DocumentClass string  `json:"documentClass"` // article, report, letter, book
	PaperSize     string  `json:"paperSize"`     // a4paper, letterpaper, legalpaper
	FontSize      string  `json:"fontSize"`      // 10pt, 11pt, 12pt
	FontFamily    string  `json:"fontFamily"`    // default, helvet, times, palatino, courier
	MarginTop     string  `json:"marginTop"`     // e.g. "0.5in"
	MarginBottom  string  `json:"marginBottom"`
	MarginLeft    string  `json:"marginLeft"`
	MarginRight   string  `json:"marginRight"`
	LineSpacing   float64 `json:"lineSpacing"` // 1.0, 1.15, 1.5, 2.0
	Alignment     string  `json:"alignment"`   // left, center, justify
	HeaderText    string  `json:"headerText"`
	FooterText    string  `json:"footerText"`
}

// DefaultPageSettings returns sensible defaults for document layout.
func DefaultPageSettings() PageSettings {
	return PageSettings{
		DocumentClass: "article",
		PaperSize:     "a4paper",
		FontSize:      "10pt",
		FontFamily:    "default",
		MarginTop:     "0.60in",
		MarginBottom:  "0.55in",
		MarginLeft:    "0.70in",
		MarginRight:   "0.70in",
		LineSpacing:   1.0,
		Alignment:     "justify",
	}
}

// MergeDefaults fills empty PageSettings fields with defaults.
func (ps PageSettings) MergeDefaults() PageSettings {
	d := DefaultPageSettings()
	if ps.DocumentClass == "" {
		ps.DocumentClass = d.DocumentClass
	}
	if ps.PaperSize == "" {
		ps.PaperSize = d.PaperSize
	}
	if ps.FontSize == "" {
		ps.FontSize = d.FontSize
	}
	if ps.FontFamily == "" {
		ps.FontFamily = d.FontFamily
	}
	if ps.MarginTop == "" {
		ps.MarginTop = d.MarginTop
	}
	if ps.MarginBottom == "" {
		ps.MarginBottom = d.MarginBottom
	}
	if ps.MarginLeft == "" {
		ps.MarginLeft = d.MarginLeft
	}
	if ps.MarginRight == "" {
		ps.MarginRight = d.MarginRight
	}
	if ps.LineSpacing == 0 {
		ps.LineSpacing = d.LineSpacing
	}
	if ps.Alignment == "" {
		ps.Alignment = d.Alignment
	}
	return ps
}

// RenderData is passed to Go templates — combines CV content with page settings.
type RenderData struct {
	CVData
	Settings PageSettings
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

// fmtSpacing returns the LaTeX command for the given line spacing value.
func fmtSpacing(v float64) string {
	switch {
	case v <= 1.0:
		return `\singlespacing`
	case v == 1.5:
		return `\onehalfspacing`
	case v >= 2.0:
		return `\doublespacing`
	default:
		return fmt.Sprintf(`\setstretch{%.2f}`, v)
	}
}

// fontPkg returns the LaTeX usepackage line(s) for the selected font family.
func fontPkg(family string) string {
	switch family {
	case "helvet":
		return "\\usepackage{helvet}\n\\renewcommand{\\familydefault}{\\sfdefault}"
	case "times":
		return "\\usepackage{mathptmx}"
	case "palatino":
		return "\\usepackage{palatino}"
	case "courier":
		return "\\usepackage{courier}\n\\renewcommand{\\familydefault}{\\ttdefault}"
	default:
		return "% default Computer Modern font"
	}
}

// Render processes the given LaTeX template with the provided CVData and PageSettings.
func Render(name string, data CVData, ps PageSettings) (string, error) {
	raw, err := Get(name)
	if err != nil {
		return "", err
	}

	ps = ps.MergeDefaults()

	t := template.New(name).Delims("[[", "]]").Funcs(template.FuncMap{
		"texEscape":  texEscape,
		"texLines":   texLines,
		"fmtSpacing": fmtSpacing,
		"fontPkg":    fontPkg,
		"sprintf":    fmt.Sprintf,
		"ne": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
		},
	})

	t, err = t.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	rd := RenderData{
		CVData:   data,
		Settings: ps,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, rd); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
