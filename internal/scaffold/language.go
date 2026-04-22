package scaffold

import (
	"fmt"
	"strings"
)

type Language string

const (
	LanguageEnglish Language = "en"
	LanguageChinese Language = "zh"
)

type Template struct {
	Language  Language
	Label     string
	LocalName string
	RemoteURL string
}

var templates = map[Language]Template{
	LanguageEnglish: {
		Language:  LanguageEnglish,
		Label:     "English",
		LocalName: "harness-template",
		RemoteURL: "https://github.com/iFurySt/harness-template.git",
	},
	LanguageChinese: {
		Language:  LanguageChinese,
		Label:     "Chinese",
		LocalName: "harness-template-cn",
		RemoteURL: "https://github.com/iFurySt/harness-template-cn.git",
	},
}

func ParseLanguage(value string) (Language, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "1", "en", "eng", "english":
		return LanguageEnglish, nil
	case "2", "zh", "cn", "zh-cn", "chinese", "mandarin":
		return LanguageChinese, nil
	default:
		return "", fmt.Errorf("unsupported language %q; use en or zh", value)
	}
}

func TemplateFor(language Language) (Template, error) {
	tpl, ok := templates[language]
	if !ok {
		return Template{}, fmt.Errorf("unsupported language %q", language)
	}
	return tpl, nil
}
