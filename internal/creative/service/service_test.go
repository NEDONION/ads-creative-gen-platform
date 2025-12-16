package service

import (
	"strings"
	"testing"
)

func TestGeneratePromptIncludesTitleAndSellingPoints(t *testing.T) {
	svc := &CreativeService{}
	title := "Test Product"
	selling := []string{"fast", "light"}
	style := "modern"

	prompt := svc.generatePrompt(title, selling, style)
	if !strings.Contains(prompt, title) {
		t.Fatalf("prompt missing title: %s", prompt)
	}
	for _, s := range selling {
		if !strings.Contains(prompt, s) {
			t.Fatalf("prompt missing selling point %s: %s", s, prompt)
		}
	}
	if !strings.Contains(prompt, style) {
		t.Fatalf("prompt missing style: %s", prompt)
	}
}
