package main

import (
"encoding/json"
"fmt"
"log"

"github.com/semmidev/atstex-lab/internal/cvtemplate"
"github.com/semmidev/atstex-lab/internal/builder"
)

func main() {
	var cvData cvtemplate.CVData
	
	// Simulating user db string
	biodata := []byte(`{"skills": "Go, Python, Java"}`)
	err := json.Unmarshal(biodata, &cvData)
	fmt.Println("Unmarshal error:", err)
	
	configRaw := []byte(`{
		"theme_color": "#000000",
		"columns": 1,
		"layout": ["header", "summary", "experience", "education", "skills", "projects", "certifications", "languages"]
	}`)
	
	tex, err := builder.Generate(configRaw, cvData)
	if err != nil {
		log.Fatal("Generate error:", err)
	}
	fmt.Println("Tex len:", len(tex))
}
