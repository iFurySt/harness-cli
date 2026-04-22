package scaffold

import "testing"

func TestParseLanguage(t *testing.T) {
	tests := map[string]Language{
		"":        LanguageEnglish,
		"1":       LanguageEnglish,
		"english": LanguageEnglish,
		"EN":      LanguageEnglish,
		"2":       LanguageChinese,
		"zh":      LanguageChinese,
		"cn":      LanguageChinese,
		"zh-CN":   LanguageChinese,
	}

	for input, want := range tests {
		got, err := ParseLanguage(input)
		if err != nil {
			t.Fatalf("ParseLanguage(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseLanguage(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseLanguageRejectsUnknown(t *testing.T) {
	if _, err := ParseLanguage("fr"); err == nil {
		t.Fatal("ParseLanguage(fr) returned nil error")
	}
}
