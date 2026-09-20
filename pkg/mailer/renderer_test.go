package mailer

import (
	"strings"
	"testing"
)

func TestRender_EscapesUserInput(t *testing.T) {
	const xss = `<script>alert(1)</script>`
	const evilLink = `javascript:alert(1)`

	body, err := Render(NewApplicant, NewApplicantsFormInput{
		VacancyName: xss,
		Name:        xss,
		Email:       xss,
		Phone:       xss,
		City:        xss,
		Experience:  xss,
		AboutSelf:   xss,
		Link:        evilLink,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if strings.Contains(body, "<script>") {
		t.Errorf("Render() did not escape HTML input, got raw <script> tag in body:\n%s", body)
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Errorf("Render() expected escaped script tag in body, got:\n%s", body)
	}
	if strings.Contains(body, `href="javascript:alert(1)"`) {
		t.Errorf("Render() allowed a javascript: URL through into href, got:\n%s", body)
	}
}

func TestRender_NewPlanEscapesUserInput(t *testing.T) {
	const xss = `"><img src=x onerror=alert(1)>`

	body, err := Render(NewPlan, NewPlanFormInput{
		Name:        xss,
		Email:       xss,
		Direction:   xss,
		Description: xss,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if strings.Contains(body, "<img src=x onerror=alert(1)>") {
		t.Errorf("Render() did not escape HTML input, got raw <img> tag in body:\n%s", body)
	}
}

func TestRender_StaticTemplatesWithNilData(t *testing.T) {
	for _, name := range []string{NotifyAboutPlan, NotifyAboutVacancy} {
		if _, err := Render(name, nil); err != nil {
			t.Errorf("Render(%q, nil) error = %v", name, err)
		}
	}
}

func TestRender_ReusesCachedTemplate(t *testing.T) {
	before := templates.Lookup(NewApplicant)
	if before == nil {
		t.Fatal("expected template to be pre-parsed and cached at package init")
	}

	if _, err := Render(NewApplicant, NewApplicantsFormInput{Name: "test"}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	after := templates.Lookup(NewApplicant)
	if before != after {
		t.Error("expected Render() to reuse the cached parsed template, got a different instance")
	}
}
