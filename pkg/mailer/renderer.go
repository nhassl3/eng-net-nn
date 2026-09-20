package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

const (
	NewApplicant       = "new_applicant.html"
	NewPlan            = "new_plan.html"
	NotifyAboutPlan    = "notify_user_about_plan.html"
	NotifyAboutVacancy = "notify_user_about_vacancy.html"
)

//go:embed templates/*.html
var templatesFS embed.FS

// templates holds every template parsed once at package init, keyed by file name.
var templates = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

func Render(name string, data interface{}) (string, error) {
	if !strings.HasSuffix(name, ".html") {
		name = name + ".html"
	}

	var body bytes.Buffer
	if err := templates.ExecuteTemplate(&body, name, data); err != nil {
		return "", fmt.Errorf("renderer.Render: failed to execute template: %w", err)
	}

	return body.String(), nil
}
