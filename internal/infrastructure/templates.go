package infrastructure

import (
	"bytes"
	"strings"
	"text/template"
)

// GoTemplateEngine implements domain.TemplatePort using text/template
type GoTemplateEngine struct{}

func NewGoTemplateEngine() *GoTemplateEngine {
	return &GoTemplateEngine{}
}

func (t *GoTemplateEngine) Render(name, tmpl string, data interface{}) ([]byte, error) {
	funcMap := template.FuncMap{
		"title": func(s string) string {
			if s == "" {
				return ""
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"pascal": func(s string) string {
			if s == "" {
				return ""
			}
			parts := strings.Split(s, "_")
			for i := range parts {
				if parts[i] == "" {
					continue
				}
				parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
			}
			return strings.Join(parts, "")
		},
	}

	tObj, err := template.New(name).Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tObj.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
