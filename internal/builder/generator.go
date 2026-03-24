package builder

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/semmidev/atstex-lab/internal/cvtemplate"
)

// Config represents the raw JSON layout payload sent from the UI builder.
type Config struct {
	ThemeColor string   `json:"theme_color"`
	Columns    int      `json:"columns"`
	Layout     []string `json:"layout"`
}

// Generate takes the user layout configuration and CVData to produce a customized LaTeX document structure
// It borrows the beautiful aesthetic layout systems and macros from sky.tex
func Generate(configJSON json.RawMessage, data cvtemplate.CVData) (string, error) {
	var cfg Config
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return "", err
	}

	// Normalize color if absent
	themeColor := cfg.ThemeColor
	if themeColor == "" {
		themeColor = "26, 115, 191" // Default skyblue RGB
	} else if strings.HasPrefix(themeColor, "#") {
		// Quick hex to RGB conversion for LaTeX
		hex := strings.TrimPrefix(themeColor, "#")
		if len(hex) == 6 {
			// e.g. RGB {26, 115, 191}
			// In production would properly convert hex to decimal, but let's just emit raw hex color for xcolor
			themeColor = hex
		}
	}

	blocks := map[string]string{
		"header": `
% --- HEADER ---
\begin{center}
{{if .Personal.Photo}}
  \begin{minipage}[c]{0.15\textwidth}
    \centering
    \begin{tikzpicture}
      \begin{scope}
        \clip (0,0) circle (1.2cm);
        \node[anchor=center] at (0,0)
          {\includegraphics[width=2.4cm,height=2.4cm]{photo}};
      \end{scope}
      \draw[skyblue, line width=1.2pt] (0,0) circle (1.2cm);
    \end{tikzpicture}
  \end{minipage}%
  \hspace{0.03\textwidth}%
  \begin{minipage}[c]{0.78\textwidth}
    \raggedright
    {\Huge\bfseries\scshape\color{skyblue} {{- texEscape .Personal.Name -}} }%
    {{if .Personal.Title}}\\\vspace{2pt}{\large\color{darkgrey} {{- texEscape .Personal.Title -}} }{{end}}
    \\\vspace{3pt}
    {{if .Personal.Location}}{\small\color{midgrey} {{- texEscape .Personal.Location -}} }\\[2pt]{{end}}
    \small
    {{if .Personal.Phone}}\href{tel:{{texEscape .Personal.Phone}}}{\color{darkgrey} {{- texEscape .Personal.Phone -}} }{{end}}%
    {{if .Personal.Email}} $|$ \href{mailto:{{texEscape .Personal.Email}}}{\color{darkgrey} {{- texEscape .Personal.Email -}} }{{end}}%
    {{if .Personal.Linkedin.URL}} $|$ \href{ {{- texEscape .Personal.Linkedin.URL -}} }{\color{skyblue} {{- if .Personal.Linkedin.Display -}} {{- texEscape .Personal.Linkedin.Display -}} {{- else -}} {{- texEscape .Personal.Linkedin.URL -}} {{- end -}} }{{end}}%
    {{if .Personal.Github.URL}} $|$ \href{ {{- texEscape .Personal.Github.URL -}} }{\color{skyblue} {{- if .Personal.Github.Display -}} {{- texEscape .Personal.Github.Display -}} {{- else -}} {{- texEscape .Personal.Github.URL -}} {{- end -}} }{{end}}%
    {{if .Personal.Website.URL}} $|$ \href{ {{- texEscape .Personal.Website.URL -}} }{\color{skyblue} {{- if .Personal.Website.Display -}} {{- texEscape .Personal.Website.Display -}} {{- else -}} {{- texEscape .Personal.Website.URL -}} {{- end -}} }{{end}}
  \end{minipage}
{{else}}
  %% No-photo variant — centred name + inline contacts
  {\Huge\bfseries\scshape\color{skyblue} {{- texEscape .Personal.Name -}} }%
  {{if .Personal.Title}}\\\vspace{2pt}{\large\color{darkgrey} {{- texEscape .Personal.Title -}} }{{end}}
  \\\vspace{3pt}
  {{if .Personal.Location}}{\small\color{midgrey} {{- texEscape .Personal.Location -}} }\\[2pt]{{end}}
  \small
  {{if .Personal.Phone}}\href{tel:{{texEscape .Personal.Phone}}}{\color{darkgrey} {{- texEscape .Personal.Phone -}} }{{end}}%
  {{if .Personal.Email}} $|$ \href{mailto:{{texEscape .Personal.Email}}}{\color{darkgrey} {{- texEscape .Personal.Email -}} }{{end}}%
  {{if .Personal.Linkedin.URL}} $|$ \href{ {{- texEscape .Personal.Linkedin.URL -}} }{\color{skyblue} {{- if .Personal.Linkedin.Display -}} {{- texEscape .Personal.Linkedin.Display -}} {{- else -}} {{- texEscape .Personal.Linkedin.URL -}} {{- end -}} }{{end}}%
  {{if .Personal.Github.URL}} $|$ \href{ {{- texEscape .Personal.Github.URL -}} }{\color{skyblue} {{- if .Personal.Github.Display -}} {{- texEscape .Personal.Github.Display -}} {{- else -}} {{- texEscape .Personal.Github.URL -}} {{- end -}} }{{end}}%
  {{if .Personal.Website.URL}} $|$ \href{ {{- texEscape .Personal.Website.URL -}} }{\color{skyblue} {{- if .Personal.Website.Display -}} {{- texEscape .Personal.Website.Display -}} {{- else -}} {{- texEscape .Personal.Website.URL -}} {{- end -}} }{{end}}
{{end}}
\end{center}
\vspace{-2pt}
`,
		"summary": `
{{if .Summary}}
\section{Professional Summary}
\skyListStart
  \item {\small\color{darkgrey} {{- texEscape .Summary -}} }
\skyListEnd
\vspace{2pt}
{{end}}
`,
		"experience": `
{{if .Experience}}
\section{Professional Experience}
\skyListStart
  {{range .Experience}}
  \skySubheading
    { {{- texEscape .Company -}} }{ {{- if .Location -}} {{- texEscape .Location -}} {{- end -}} }
    { {{- texEscape .Title -}} }{ {{- texEscape .Dates -}} }
    {{if .Bullets}}
    \skyBulletStart
      {{range texLines .Bullets}}\skyItem{ {{- texEscape . -}} }
      {{end}}
    \skyBulletEnd
    {{end}}
  {{end}}
\skyListEnd
{{end}}
`,
		"education": `
{{if .Education}}
\section{Education}
\skyListStart
  {{range .Education}}
  \skySubheading
    { {{- texEscape .Institution -}} }{ {{- if .Location -}} {{- texEscape .Location -}} {{- end -}} }
    { {{- texEscape .Degree -}} {{if .GPA}} {\normalfont --} \textbf{GPA: {{texEscape .GPA}}}{{end}}}{ {{- texEscape .Dates -}} }
    {{if .Activities}}
    \skyBulletStart
      \skyItem{\textit{Activities \& Societies:} {{- texEscape .Activities -}} }
    \skyBulletEnd
    {{end}}
  {{end}}
\skyListEnd
{{end}}
`,
		"skills": `
{{if or .Skills.Languages .Skills.Frameworks .Skills.Tools .Skills.Other}}
\section{Technical Skills}
\vspace{-2pt}
\begin{tabular}{@{}p{3cm}@{}p{\dimexpr\linewidth-3cm\relax}@{}}
  {{if .Skills.Languages}}\small\textbf{\color{darkgrey}Languages}    & \small{\color{darkgrey} {{- texEscape .Skills.Languages -}} } \\[3pt]{{end}}
  {{if .Skills.Frameworks}}\small\textbf{\color{darkgrey}Technologies} & \small{\color{darkgrey} {{- texEscape .Skills.Frameworks -}} } \\[3pt]{{end}}
  {{if .Skills.Tools}}\small\textbf{\color{darkgrey}Tools}         & \small{\color{darkgrey} {{- texEscape .Skills.Tools -}} } \\[3pt]{{end}}
  {{if .Skills.Other}}\small\textbf{\color{darkgrey}Other}         & \small{\color{darkgrey} {{- texEscape .Skills.Other -}} } \\{{end}}
\end{tabular}
\vspace{2pt}
{{end}}
`,
		"projects": `
{{if .Projects}}
\section{Selected Projects}
\skyListStart
  {{range .Projects}}
  \skyProjectRow{%
    \textbf{\color{darkgrey} {{- texEscape .Name -}} }%
    {{if .Role}} $|$ {\emph{\color{darkgrey} {{- texEscape .Role -}} }}{{end}}%
    {{if .Link}} $|$ \href{https://{{texEscape .Link}}}{ {\small\color{skyblue}\texttt{ {{- texEscape .Link -}} }} }{{end}}%
  }{}
  {{if .Bullets}}
  \skyBulletStart
    {{range texLines .Bullets}}\skyItem{ {{- texEscape . -}} }
    {{end}}
  \skyBulletEnd
  {{end}}
  {{end}}
\skyListEnd
{{end}}
`,
		"certifications": `
{{if .Certifications}}
\section{Certifications \& Courses}
\vspace{4pt}
{{range .Certifications}}
\noindent\begin{tabular*}{\linewidth}{@{}p{0.85\linewidth}@{\extracolsep{\fill}}r@{}}
  \small\textbf{\color{darkgrey} {{- texEscape .Name -}} }, {\color{darkgrey} {{- texEscape .Issuer -}} } & \\
\end{tabular*}\vspace{6pt}
{{end}}
{{end}}
`,
		"languages": `
{{if .Skills.Languages}}
\section{Bahasa}
\vspace{2pt}
\small{\color{darkgrey} {{- texEscape .Skills.Languages -}} }
\vspace{4pt}
{{end}}
`,
	}

	// Document preamble with `sky.tex` styling
	doc := `
\documentclass[10pt,a4paper]{article}
\usepackage[T1]{fontenc}
\usepackage[utf8]{inputenc}
\usepackage[english]{babel}
\usepackage[top=0.7in, bottom=0.7in, left=0.7in, right=0.7in]{geometry}
\usepackage{array}
\usepackage{tabularx}
\usepackage{booktabs}
\usepackage{enumitem}
\usepackage[usenames,dvipsnames]{xcolor}
\usepackage{graphicx}
\usepackage{microtype}
\usepackage{titlesec}
\usepackage{etoolbox}
\usepackage{latexsym}
\usepackage{marvosym}
\usepackage{amssymb}
\usepackage{verbatim}
\usepackage{tikz}
\usepackage{multicol}
\usepackage[hidelinks]{hyperref}

% Colors based on sky.tex
\definecolor{skyblue}{HTML}{1A73BF} % Reusing a solid hex by default or could be dynamic
\definecolor{skylight}{RGB}{232, 244, 253}
\definecolor{darkgrey}{RGB}{60, 60, 60}
\definecolor{midgrey}{RGB}{110, 110, 110}

\hypersetup{colorlinks=false, urlcolor=skyblue, linkcolor=skyblue}

\raggedright
\raggedbottom
\setlength{\parindent}{0pt}
\setlength{\parskip}{0pt}
\setlength{\tabcolsep}{0in}
\urlstyle{same}

\titleformat{\section}{%
  \vspace{-2pt}\large\bfseries\scshape\color{skyblue}\raggedright%
}{}{0em}{}[{%
  \color{skyblue}\titlerule[0.8pt]%
}\vspace{-2pt}]
\titlespacing{\section}{0pt}{10pt}{6pt}

\newcommand{\skySubheading}[4]{%
  \vspace{2pt}\item
  \begin{tabular*}{\linewidth}{@{}l@{\extracolsep{\fill}}r@{}}
    \textbf{\color{darkgrey}#1} & \small\textit{\color{midgrey}#2} \\
    \textit{\small\color{darkgrey}#3} & \textit{\small\color{midgrey}#4} \\
  \end{tabular*}\vspace{-2pt}%
}

\newcommand{\skyProjectRow}[2]{%
  \vspace{2pt}\item
  \begin{tabular*}{\linewidth}{@{}l@{\extracolsep{\fill}}r@{}}
    \small #1 & \small\textit{\color{midgrey}#2} \\
  \end{tabular*}\vspace{-2pt}%
}

\newcommand{\skyItem}[1]{\item {\small\color{darkgrey}#1}\vspace{0pt}}
\newcommand{\skyListStart}{\begin{itemize}[leftmargin=0.15in, label={}]}
\newcommand{\skyListEnd}{\end{itemize}}
\newcommand{\skyBulletStart}{\begin{itemize}[leftmargin=0.20in, label={\textcolor{skyblue}{\tiny$\blacksquare$}}, itemsep=1pt, topsep=3pt, parsep=1pt]}
\newcommand{\skyBulletEnd}{\end{itemize}\vspace{-2pt}}

\begin{document}
`

	headerIndex := -1
	if cfg.Columns == 2 && len(cfg.Layout) > 0 {
		// Render header outside multicolumn if it's the very first block
		if cfg.Layout[0] == "header" {
			headerIndex = 0
			doc += blocks["header"]
		}
		doc += "\\begin{multicols}{2}\n"
	}

	for i, blockType := range cfg.Layout {
		if i == headerIndex {
			continue
		}
		if content, exists := blocks[blockType]; exists {
			doc += content
		}
	}

	if cfg.Columns == 2 {
		doc += "\\end{multicols}\n"
	}

	doc += `
\end{document}
`

	// Compile the resulting string as a Go text/template with the user's data
	funcs := template.FuncMap{
		"join": strings.Join,
		"texLines": func(s string) []string {
			return strings.Split(s, "\n")
		},
		"texEscape": func(s string) string {
			replacer := strings.NewReplacer(
				"&", "\\&", "%", "\\%", "$", "\\$", "#", "\\#", "_", "\\_", "{", "\\{", "}", "\\}",
				"~", "\\textasciitilde{}", "^", "\\textasciicircum{}", "\\", "\\textbackslash{}",
			)
			return replacer.Replace(s)
		},
	}

	tmpl, err := template.New("custom").Funcs(funcs).Parse(doc)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
