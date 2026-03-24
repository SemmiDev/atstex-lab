package main

import (
"log"
"os"
"text/template"
"github.com/semmidev/atstex-lab/internal/cvtemplate"
)

func main() {
	doc := `{{if .Skills}} {{.Skills.Languages}} {{end}}`
	tmpl, err := template.New("custom").Parse(doc)
	if err != nil { log.Fatal(err) }
	
	data := cvtemplate.CVData{}
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		log.Fatal(err)
	}
}
