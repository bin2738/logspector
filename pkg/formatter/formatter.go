package formatter

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type Formatter struct {
	mu       sync.RWMutex
	jsonFlag bool
	noColor  bool
}

func NewFormatter(jsonFlag, noColor bool) *Formatter {
	return &Formatter{
		jsonFlag: jsonFlag,
		noColor:  noColor,
	}
}

func (f *Formatter) Update(jsonFlag, noColor bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.jsonFlag = jsonFlag
	f.noColor = noColor
}

func (f *Formatter) Format(line string, prefix string) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if line == "" {
		return
	}

	var prefixStr string
	if prefix != "" {
		if f.noColor {
			prefixStr = fmt.Sprintf("[%s] ", prefix)
		} else {
			prefixStr = color.New(color.FgCyan).Sprintf("[%s] ", prefix)
		}
	}

	// Если включён JSON-режим, пытаемся распарсить строку
	if f.jsonFlag {
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(line), &js); err == nil {
			output, _ := json.MarshalIndent(js, "", "  ")
			fmt.Printf("%s%s\n", prefixStr, string(output))
			return
		}
		// Если err != nil, значит это не JSON, идем дальше к обычному цветному выводу
	}

	if f.noColor {
		fmt.Printf("%s%s\n", prefixStr, line)
		return
	}

	// Цветной вывод
	level := detectLevel(line)
	var coloredLine string
	switch level {
	case "ERROR":
		coloredLine = color.New(color.FgRed).Sprint(line)
	case "WARN":
		coloredLine = color.New(color.FgYellow).Sprint(line)
	case "INFO":
		coloredLine = color.New(color.FgGreen).Sprint(line)
	case "DEBUG":
		coloredLine = color.New(color.FgBlue).Sprint(line)
	default:
		coloredLine = line
	}

	fmt.Printf("%s%s\n", prefixStr, coloredLine)
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
