package parser

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type Parser struct {
	mu           sync.RWMutex
	levelFilter  string
	queryFilter  string
	invertFilter string
	sinceFilter  time.Time
	untilFilter  time.Time
}

func NewParser(level, query, invert string, since, until time.Time) *Parser {
	return &Parser{
		levelFilter:  level,
		queryFilter:  query,
		invertFilter: invert,
		sinceFilter:  since,
		untilFilter:  until,
	}
}

func (p *Parser) Update(level, query, invert string, since, until time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.levelFilter = level
	p.queryFilter = query
	p.invertFilter = invert
	p.sinceFilter = since
	p.untilFilter = until
}

func (p *Parser) Parse(line string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.invertFilter != "" && strings.Contains(line, p.invertFilter) {
		return ""
	}
	if p.queryFilter != "" && !strings.Contains(line, p.queryFilter) {
		return ""
	}

	// Фильтрация по времени (если флаги заданы и удалось извлечь время из строки)
	if !p.sinceFilter.IsZero() || !p.untilFilter.IsZero() {
		if t, ok := extractTime(line); ok {
			if !p.sinceFilter.IsZero() && t.Before(p.sinceFilter) {
				return "" // Строка старше, чем нужно
			}
			if !p.untilFilter.IsZero() && t.After(p.untilFilter) {
				return "" // Строка новее, чем нужно
			}
		}
	}

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

// extractTime пытается найти временную метку в начале строки или внутри JSON
func extractTime(line string) (time.Time, bool) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	}

	// 1. Поиск внутри JSON
	if IsJSON(line) {
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(line), &js); err == nil {
			for _, key := range []string{"time", "timestamp", "@timestamp", "date"} {
				if val, ok := js[key].(string); ok {
					for _, f := range formats {
						if t, err := time.Parse(f, val); err == nil {
							return t, true
						}
					}
				}
			}
		}
	}

	// 2. Поиск в текстовой строке (обычно дата идет первой)
	cleanLine := strings.TrimPrefix(strings.TrimSpace(line), "[") // На случай формата [2023-01-01 ...]
	parts := strings.SplitN(cleanLine, " ", 3)

	// Проверяем первое слово
	if len(parts) > 0 {
		for _, f := range formats {
			if t, err := time.Parse(f, parts[0]); err == nil {
				return t, true
			}
		}
	}
	// Проверяем первые два слова (например, для "2006-01-02 15:04:05")
	if len(parts) > 1 {
		str := strings.TrimSuffix(parts[0]+" "+parts[1], "]")
		for _, f := range formats {
			if t, err := time.Parse(f, str); err == nil {
				return t, true
			}
		}
	}

	return time.Time{}, false
}
