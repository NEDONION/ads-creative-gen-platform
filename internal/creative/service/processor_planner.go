package service

import (
	"strings"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/settings"
)

type GenRequest struct {
	VariantIndex int
	Prompt       string
	Style        string
	Format       string
	Size         string
	NumImages    int
}

func (p *TaskProcessor) buildPlan(task *models.CreativeTask) []GenRequest {
	numVariants := task.NumVariants
	if numVariants <= 0 {
		numVariants = 1
	}

	if !p.hasVariantPlan(task) {
		style := styleAt(task.RequestedStyles, 0)
		prompt := generatePrompt(task.Title, task.SellingPoints, style)
		format := formatAt(task.RequestedFormats, 0, settings.DefaultFormat)

		return []GenRequest{{
			VariantIndex: 0,
			Prompt:       prompt,
			Style:        style,
			Format:       format,
			Size:         settings.DefaultImageSize,
			NumImages:    numVariants,
		}}
	}

	plan := make([]GenRequest, 0, numVariants)
	for idx := 0; idx < numVariants; idx++ {
		style := styleAt(task.VariantStyles, idx)
		if style == "" {
			style = styleAt(task.RequestedStyles, 0)
		}

		prompt := strings.TrimSpace(styleAt(task.VariantPrompts, idx))
		if prompt == "" {
			prompt = generatePrompt(task.Title, task.SellingPoints, style)
		}

		plan = append(plan, GenRequest{
			VariantIndex: idx,
			Prompt:       prompt,
			Style:        style,
			Format:       formatAt(task.RequestedFormats, idx, settings.DefaultFormat),
			Size:         settings.DefaultImageSize,
			NumImages:    1,
		})
	}

	return plan
}

func (p *TaskProcessor) hasVariantPlan(task *models.CreativeTask) bool {
	return len(task.VariantPrompts) > 0 || len(task.VariantStyles) > 1
}

func formatAt(formats []string, idx int, defaultFormat string) string {
	if len(formats) > idx && formats[idx] != "" {
		return formats[idx]
	}
	if len(formats) > 0 && formats[0] != "" {
		return formats[0]
	}
	return defaultFormat
}
