package cvtemplate

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

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
				Description: fmt.Sprintf("%s CV Template", strings.ToTitle(name)),
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
