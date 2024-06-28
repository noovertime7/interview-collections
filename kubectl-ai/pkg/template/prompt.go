package template

import (
	"bytes"
	"fmt"
	"html/template"
)

const DiagnoseTpl string = `
You are a professional kubernetes administrator.
You inspect the object and find out what might cause the error.
If there is no error, say "Everything is OK".
Write down the possible causes in bullet points, using the imperative tense.
Answer in {{ .Lang }}.

THE OBJECT:
'''
{{ .Data -}}
'''

Remember to write only the most important points and do not write more than a few bullet points.

The cause of the error might be:
`

type PromptData struct {
	Data string
	Lang string
}

func NewData(data string, lang string) *PromptData {
	return &PromptData{Data: data, Lang: lang}
}

func (this *PromptData) Parse(text string) ([]byte, error) {
	tpl := template.New("promptTpl")
	parse, err := tpl.Parse(text)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	reader := &bytes.Buffer{}
	_ = parse.Execute(reader, this)
	return reader.Bytes(), nil
}
