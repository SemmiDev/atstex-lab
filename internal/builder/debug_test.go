package builder

import (
	"fmt"
	"strings"
	"testing"
	"github.com/semmidev/atstex-lab/internal/cvtemplate"
)

func TestGeneratorDebug(t *testing.T) {
	layoutJSON := []byte(`{"columns": 1, "layout": ["header", "summary", "experience", "education", "skills", "projects", "certifications", "languages"]}`)
	var data cvtemplate.CVData
	
	_, err := Generate(layoutJSON, data)
	if err != nil {
		fmt.Printf("ERROR CAUGHT: %v\n", err)
		if strings.Contains(err.Error(), "unexpected") {
			t.Fatalf("Failed to parse: %v", err)
		}
	} else {
		fmt.Println("SUCCESS")
	}
}
