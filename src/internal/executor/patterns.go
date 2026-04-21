package executor

import (
	"fmt"
	"regexp"
	"strings"

	"makedo/internal/nodes"
)

var TypePatterns = map[string]string{
	"date":    `\d{4}-\d{2}-\d{2}`,
	"time":    `\d{2}:\d{2}:\d{2}`,
	"uuid":    `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
	"ip":      `\d{1,3}(?:\.\d{1,3}){3}`,
	"number":  `[-+]?\d*\.?\d+`,
	"version": `\d+\.\d+\.\d+(?:\.\d+)?`,
}

var typePatternRegex = regexp.MustCompile(`\$\{\{\s*([a-zA-Z0-9_-]+)\s*\}\}`)

// PrecompileDirectives processes and precompiles regex patterns for out and outr directives.
// It returns a map of compiled regexes, or an error if a regex is invalid or an unknown type is used.
func PrecompileDirectives(directives []*nodes.Directive, source []byte) (map[*nodes.Directive]*regexp.Regexp, error) {
	compiledPatterns := make(map[*nodes.Directive]*regexp.Regexp)

	for _, d := range directives {
		content := d.ContentString(source)

		switch d.Kind {
		case nodes.DirectiveOutRegex:
			// Replace ${{type}} with regex pattern
			pattern, err := expandTypes(content)
			if err != nil {
				return nil, err
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex '%s': %w", pattern, err)
			}
			compiledPatterns[d] = re

		case nodes.DirectiveOut:
			// Auto-upgrade to regex if it contains ${{type}}
			if typePatternRegex.MatchString(content) {
				pattern, err := expandTypesWithEscaping(content)
				if err != nil {
					return nil, err
				}
				re, err := regexp.Compile(pattern)
				if err != nil {
					return nil, fmt.Errorf("invalid regex '%s' generated from out directive: %w", pattern, err)
				}
				compiledPatterns[d] = re
			}
		}
	}

	return compiledPatterns, nil
}

func expandTypes(content string) (string, error) {
	var err error
	expanded := typePatternRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := typePatternRegex.FindStringSubmatch(match)
		if len(matches) > 1 {
			typeName := matches[1]
			if pattern, ok := TypePatterns[typeName]; ok {
				return pattern
			}
			err = fmt.Errorf("unknown type '%s'", typeName)
			return match // Return original on error, will be caught by err check
		}
		return match
	})
	return expanded, err
}

func expandTypesWithEscaping(content string) (string, error) {
	var err error

	// Split the content by the type pattern matches
	matches := typePatternRegex.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return regexp.QuoteMeta(content), nil
	}

	var sb strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start := match[0]
		end := match[1]

		// Escape literal part before the match
		if start > lastIndex {
			sb.WriteString(regexp.QuoteMeta(content[lastIndex:start]))
		}

		// Extract type name and replace with pattern
		typeMatch := typePatternRegex.FindStringSubmatch(content[start:end])
		if len(typeMatch) > 1 {
			typeName := typeMatch[1]
			if pattern, ok := TypePatterns[typeName]; ok {
				sb.WriteString(pattern)
			} else {
				err = fmt.Errorf("unknown type '%s'", typeName)
				sb.WriteString(content[start:end]) // Return original on error
			}
		}

		lastIndex = end
	}

	// Escape literal part after the last match
	if lastIndex < len(content) {
		sb.WriteString(regexp.QuoteMeta(content[lastIndex:]))
	}

	return sb.String(), err
}
