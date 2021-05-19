// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/syntax/rules"
)

var language = flag.String("language", "", "If provided, generate only this language")

func main() {
	flag.Parse()

	specs := []LanguageSpec{
		{
			Name:  "Plaintext",
			Rules: rules.PlaintextRules(),
		},
		{
			Name:  "Json",
			Rules: rules.JsonRules(),
		},
		{
			Name:  "Go",
			Rules: rules.GolangRules(),
		},
	}

	filteredSpecs := filterSpecs(specs, *language)
	if len(filteredSpecs) == 0 {
		fmt.Printf("No matching languages found\n")
		os.Exit(1)
	}

	if err := generateLanguageIndex(specs, "languages.go"); err != nil {
		log.Fatalf("Error generating language index: %v\n", err)
	}

	for _, spec := range filteredSpecs {
		err := generateTokenizer(
			spec.TokenizerName(),
			spec.Rules,
			spec.TokenizerPath(),
		)
		if err != nil {
			log.Fatalf("Error generating tokenizer '%s': %v\n", spec.Name, err)
		}
	}
}

type LanguageSpec struct {
	Name  string
	Rules []parser.TokenizerRule
}

func (s LanguageSpec) TokenizerName() string {
	return fmt.Sprintf("%sTokenizer", s.Name)
}

func (s LanguageSpec) LanguageConst() string {
	return fmt.Sprintf("Language%s", s.Name)
}

func (s LanguageSpec) LanguageString() string {
	return strings.ToLower(s.Name)
}

func (s LanguageSpec) TokenizerPath() string {
	return fmt.Sprintf("%s_tokenizer.go", strings.ToLower(s.Name))
}

func filterSpecs(specs []LanguageSpec, filter string) []LanguageSpec {
	if filter == "" {
		return specs
	}

	filteredSpecs := make([]LanguageSpec, 0, len(specs))
	for _, spec := range specs {
		if spec.LanguageString() == filter {
			filteredSpecs = append(filteredSpecs, spec)
		}
	}
	return filteredSpecs
}

func generateLanguageIndex(specs []LanguageSpec, outputPath string) error {
	fmt.Printf("Generating %s\n", outputPath)
	f, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrapf(err, "os.Create")
	}
	defer f.Close()

	tmplStr := `// This file is generated by gen_tokenizers.go.  DO NOT EDIT.
package syntax

import (
	"log"

	"github.com/aretext/aretext/syntax/parser"
)

// Language is an enum of available languages that we can parse.
type Language int

const (
	LanguageUndefined = Language(iota)
	{{ range .Specs -}}
	{{ .LanguageConst }}
	{{ end }}
)

var AllLanguages = []Language{
	{{ range .Specs -}}
	{{ .LanguageConst }},
	{{ end }}
}

func (language Language) String() string {
	switch language {
	case LanguageUndefined:
		return "undefined"
	{{ range .Specs -}}
	case {{ .LanguageConst }}:
		return "{{ .LanguageString }}"
	{{ end -}}
	default:
		return ""
	}
}

func LanguageFromString(s string) Language {
	switch s {
	case "undefined":
		return LanguageUndefined
	{{ range .Specs -}}
	case "{{ .LanguageString }}":
		return {{ .LanguageConst }}
	{{ end -}}
	default:
		log.Printf("Unrecognized syntax language '%s'\n", s)
		return LanguageUndefined
	}
}

// TokenizerForLanguage returns a tokenizer for the specified language.
// If no tokenizer is available (e.g. for LanguageUndefined), this returns nil.
func TokenizerForLanguage(language Language) *parser.Tokenizer {
	switch language {
	{{ range .Specs -}}
	case {{ .LanguageConst }}:
		return {{ .TokenizerName }}
	{{ end -}}
	default:
		return nil
	}
}
`
	tmpl, err := template.New("root").Parse(tmplStr)
	if err != nil {
		return errors.Wrapf(err, "template.New")
	}

	return tmpl.Execute(f, map[string]interface{}{
		"Specs": specs,
	})
}

func generateTokenizer(tokenizerName string, tokenizerRules []parser.TokenizerRule, outputPath string) error {
	fmt.Printf("Generating %s\n", outputPath)
	tokenizer, err := parser.GenerateTokenizer(tokenizerRules)
	if err != nil {
		return err
	}
	return writeTokenizer(tokenizer, tokenizerName, outputPath)
}

func writeTokenizer(tokenizer *parser.Tokenizer, tokenizerName string, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrapf(err, "os.Create")
	}
	defer f.Close()

	tmplStr := `// This file is generated by gen_tokenizers.go.  DO NOT EDIT.
package syntax

import "github.com/aretext/aretext/syntax/parser"

{{ define "rule" }}
parser.TokenizerRule{
	Regexp: {{ printf "%q" .Regexp }},
	TokenRole: {{ .TokenRole }},
	SubRules: []parser.TokenizerRule{
		{{ range .SubRules -}}
		{{ template "rule" . }},
		{{ end }}
	},
}{{ end }}

{{ define "tokenizer" }}
&parser.Tokenizer{
	StateMachine: &parser.Dfa{
		NumStates: {{ .StateMachine.NumStates }},
		StartState: {{ .StateMachine.StartState }},
		Transitions: {{ printf "%#v" .StateMachine.Transitions }},
		AcceptActions: {{ printf "%#v" .StateMachine.AcceptActions }},
	},
	SubTokenizers: []*parser.Tokenizer{
		{{ range .SubTokenizers -}}
		{{ if . }}{{ template "tokenizer" . }},{{ else }}nil,{{ end }}
		{{ end }}
	},
	Rules: []parser.TokenizerRule{
		{{ range .Rules }}
		{{ template "rule" . }},
		{{ end }}
	},
}{{ end }}

var {{ .TokenizerName }} *parser.Tokenizer

func init() {
	{{ .TokenizerName }} = {{ template "tokenizer" .Tokenizer }}
}
`

	tmpl, err := template.New("root").Parse(tmplStr)
	if err != nil {
		return errors.Wrapf(err, "template.New")
	}

	return tmpl.Execute(f, map[string]interface{}{
		"TokenizerName": tokenizerName,
		"Tokenizer":     tokenizer,
	})
}
