package cvtemplate

import (
	"testing"
)

func TestRender(t *testing.T) {
	data := CVData{
		Personal: Personal{
			Name:     "John Doe",
			Email:    "john@example.com",
			Location: "Earth",
		},
		Summary: "A great engineer.",
	}

	result, err := Render("sea", data, DefaultPageSettings(), false)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if result == "" {
		t.Fatal("Result is empty")
	}

	t.Logf("Result: %s", result)
}
