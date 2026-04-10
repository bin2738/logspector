package parser

import (
	"encoding/json"
	"strings"
)

type Parser struct {
	levelFilter string
}

func NewParser(level string) *Parser {
	return &Parser{levelFilter: level}
}

func (p *Parser) Parse(line string) string {
	if p.levelFilter != "" {
		level := detectLevel(line)
		if level != p.levelFilter {
			return ""
		}
	}
	return line
}

// Простой детектор уровня лога
func detectLevel(line string) string {
	line = strings.ToUpper(line)
	switch {
	case strings.Contains(line, "ERROR"):
		return "ERROR"
	case strings.Contains(line, "WARN"):
		return "WARN"
	case strings.Contains(line, "INFO"):
		return "INFO"
	case strings.Contains(line, "DEBUG"):
		return "DEBUG"
	default:
		return ""
	}
}

func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
