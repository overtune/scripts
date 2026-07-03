// Package meta parses the embedded @meta/@end metadata block from a script file.
package meta

import (
	"bufio"
	"bytes"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// Arg describes one positional argument a script accepts.
type Arg struct {
	Name     string `yaml:"name"`
	Required bool   `yaml:"required"`
	Help     string `yaml:"help"`
}

// ScriptMeta is the parsed metadata for a single script.
type ScriptMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
	Args        []Arg  `yaml:"args"`

	Path       string `yaml:"-"` // resolved during scan
	Documented bool   `yaml:"-"` // false when no valid block was found
	Warning    string `yaml:"-"` // human-readable parse-warning message, empty when none
}

// ErrNoMetaBlock is returned when no @meta block is found in the scanned lines.
var ErrNoMetaBlock = errors.New("no metadata block found")

const maxScanLines = 20

// Parse extracts and parses the @meta/@end block from script content.
func Parse(content []byte) (ScriptMeta, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var (
		inBlock   bool
		yamlLines []string
		lineNum   int
	)
	for scanner.Scan() {
		lineNum++
		body := stripComment(scanner.Text())
		if !inBlock {
			if lineNum > maxScanLines {
				break
			}
			if strings.TrimSpace(body) == "@meta" {
				inBlock = true
			}
			continue
		}
		if strings.TrimSpace(body) == "@end" {
			var m ScriptMeta
			if err := yaml.Unmarshal([]byte(strings.Join(yamlLines, "\n")), &m); err != nil {
				return ScriptMeta{}, err
			}
			m.Documented = true
			return m, nil
		}
		yamlLines = append(yamlLines, body)
	}
	return ScriptMeta{}, ErrNoMetaBlock
}

// stripComment removes a leading # or // comment marker and one optional space,
// preserving the relative indentation of the remaining YAML content.
func stripComment(line string) string {
	s := strings.TrimLeft(line, " \t")
	switch {
	case strings.HasPrefix(s, "//"):
		s = s[2:]
	case strings.HasPrefix(s, "#"):
		s = s[1:]
	default:
		return line
	}
	return strings.TrimPrefix(s, " ")
}
