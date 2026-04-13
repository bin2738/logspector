package formatter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type Formatter struct {
	jsonFlag bool
}

func NewFormatter(jsonFlag bool) *Formatter {
	return &Formatter{jsonFlag: jsonFlag}
}

func (f *Formatter) Format(line string) {
	if line == "" {
		return
	}

	// Если включён JSON-режим, пытаемся распарсить строку
	if f.jsonFlag {
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(line), &js); err == nil {
			output, _ := json.MarshalIndent(js, "", "  ")
			fmt.Println(string(output))
			return
		}
		// Если err != nil, значит это не JSON, идем дальше к обычному цветному выводу
	}

	// Цветной вывод
	level := detectLevel(line)
	switch level {
	case "ERROR":
		color.New(color.FgRed).Println(line)
	case "WARN":
		color.New(color.FgYellow).Println(line)
	case "INFO":
		color.New(color.FgGreen).Println(line)
	case "DEBUG":
		color.New(color.FgBlue).Println(line)
	default:
		fmt.Println(line)
	}
}

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
