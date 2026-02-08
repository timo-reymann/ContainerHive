package templating

import (
	"bytes"
	"html/template"
)

type GoTemplateTemplatingProcessor struct{}

func (g *GoTemplateTemplatingProcessor) renderTemplate(path string, raw []byte, data interface{}) ([]byte, error) {
	tpl := template.New(path)
	parsed, err := tpl.Parse(string(raw))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer

	if err := parsed.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *GoTemplateTemplatingProcessor) Process(tplCtx *TemplateContext, path string, content []byte) ([]byte, error) {
	return g.renderTemplate(path, content, tplCtx)
}
