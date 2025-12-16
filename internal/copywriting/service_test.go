package copywriting

import "testing"

func TestResolveLanguage(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		lang        string
		want        string
	}{
		{"explicit zh", "任何", "zh", "zh"},
		{"explicit en", "anything", "en", "en"},
		{"auto zh", "超级好用的水杯", "", "zh"},
		{"auto en", "Portable charger", "", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveLanguage(tt.productName, tt.lang)
			if got != tt.want {
				t.Fatalf("resolveLanguage(%q,%q) = %s, want %s", tt.productName, tt.lang, got, tt.want)
			}
		})
	}
}
