// Package compiler — input sanitisation helpers.
//
// SanitizeLatexInput escapes LaTeX special characters in user-supplied strings
// so they are rendered literally and cannot inject arbitrary TeX commands.
// SanitizeCVData applies the escaper to every string field of the CV data
// struct before it is interpolated into a template.
package compiler

import (
	"strings"

	"github.com/semmidev/atstex-lab/internal/cvtemplate"
)

// latexSpecials maps every LaTeX special character to its safe replacement.
// The order matters: backslash must be replaced first so that subsequent
// replacements don't double-escape.
var latexSpecials = strings.NewReplacer(
	`\`, `\textbackslash{}`,
	`{`, `\{`,
	`}`, `\}`,
	`$`, `\$`,
	`&`, `\&`,
	`#`, `\#`,
	`^`, `\textasciicircum{}`,
	`_`, `\_`,
	`~`, `\textasciitilde{}`,
	`%`, `\%`,
)

// SanitizeLatexInput escapes all ten LaTeX special characters in s so that
// the string can be safely embedded in a LaTeX document without executing
// arbitrary commands.
func SanitizeLatexInput(s string) string {
	return latexSpecials.Replace(s)
}

// sanitizeLink escapes both fields of a Link.
func sanitizeLink(l *cvtemplate.Link) {
	l.Display = SanitizeLatexInput(l.Display)
	l.URL = SanitizeLatexInput(l.URL)
}

// SanitizeCVData applies SanitizeLatexInput to every user-supplied string
// field in the CVData struct. Call this before passing the data to a template
// renderer.
func SanitizeCVData(d *cvtemplate.CVData) {
	// ── Personal ──
	d.Personal.Name = SanitizeLatexInput(d.Personal.Name)
	d.Personal.Title = SanitizeLatexInput(d.Personal.Title)
	d.Personal.Email = SanitizeLatexInput(d.Personal.Email)
	d.Personal.Phone = SanitizeLatexInput(d.Personal.Phone)
	d.Personal.Location = SanitizeLatexInput(d.Personal.Location)
	sanitizeLink(&d.Personal.Linkedin)
	sanitizeLink(&d.Personal.Github)
	sanitizeLink(&d.Personal.Website)

	// ── Summary ──
	d.Summary = SanitizeLatexInput(d.Summary)

	// ── Experience ──
	for i := range d.Experience {
		e := &d.Experience[i]
		e.Company = SanitizeLatexInput(e.Company)
		e.Title = SanitizeLatexInput(e.Title)
		e.Location = SanitizeLatexInput(e.Location)
		e.Dates = SanitizeLatexInput(e.Dates)
		e.Bullets = SanitizeLatexInput(e.Bullets)
	}

	// ── Education ──
	for i := range d.Education {
		e := &d.Education[i]
		e.Institution = SanitizeLatexInput(e.Institution)
		e.Degree = SanitizeLatexInput(e.Degree)
		e.Location = SanitizeLatexInput(e.Location)
		e.Dates = SanitizeLatexInput(e.Dates)
		e.GPA = SanitizeLatexInput(e.GPA)
		e.Activities = SanitizeLatexInput(e.Activities)
	}

	// ── Projects ──
	for i := range d.Projects {
		p := &d.Projects[i]
		p.Name = SanitizeLatexInput(p.Name)
		p.Role = SanitizeLatexInput(p.Role)
		p.Link = SanitizeLatexInput(p.Link)
		p.Bullets = SanitizeLatexInput(p.Bullets)
	}

	// ── Skills ──
	d.Skills.Languages = SanitizeLatexInput(d.Skills.Languages)
	d.Skills.Frameworks = SanitizeLatexInput(d.Skills.Frameworks)
	d.Skills.Tools = SanitizeLatexInput(d.Skills.Tools)
	d.Skills.Other = SanitizeLatexInput(d.Skills.Other)

	// ── Certifications ──
	for i := range d.Certifications {
		c := &d.Certifications[i]
		c.Name = SanitizeLatexInput(c.Name)
		c.Issuer = SanitizeLatexInput(c.Issuer)
	}

	// ── Volunteer ──
	for i := range d.Volunteer {
		v := &d.Volunteer[i]
		v.Organization = SanitizeLatexInput(v.Organization)
		v.Role = SanitizeLatexInput(v.Role)
		v.Location = SanitizeLatexInput(v.Location)
		v.Dates = SanitizeLatexInput(v.Dates)
		v.Bullets = SanitizeLatexInput(v.Bullets)
	}

	// ── Awards ──
	for i := range d.Awards {
		a := &d.Awards[i]
		a.Title = SanitizeLatexInput(a.Title)
		a.Issuer = SanitizeLatexInput(a.Issuer)
		a.Date = SanitizeLatexInput(a.Date)
		a.Description = SanitizeLatexInput(a.Description)
	}

	// ── Talks ──
	for i := range d.Talks {
		t := &d.Talks[i]
		t.Title = SanitizeLatexInput(t.Title)
		t.Event = SanitizeLatexInput(t.Event)
		t.Location = SanitizeLatexInput(t.Location)
		t.Date = SanitizeLatexInput(t.Date)
		t.Description = SanitizeLatexInput(t.Description)
	}
}
