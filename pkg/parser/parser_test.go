package parser

import (
	"testing"
	"time"
)

// TestParser_Filters проверяет логику фильтрации по тексту и уровню логов.
func TestParser_Filters(t *testing.T) {
	// Опорное время для проверки фильтров (1 января 2023 года, 12:00:00)
	timeRef := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// Табличные тесты — задаем массив различных сценариев
	tests := []struct {
		name     string
		line     string
		level    string
		query    string
		invert   string
		since    time.Time
		until    time.Time
		expected bool // true - если строку нужно пропустить дальше, false - если скрыть
	}{
		{
			name:     "Без фильтров пропускаем все",
			line:     "INFO Application started",
			query:    "",
			invert:   "",
			expected: true,
		},
		{
			name:     "Поиск по подстроке (grep)",
			line:     "ERROR user_id=123 failed",
			query:    "user_id=123",
			invert:   "",
			expected: true,
		},
		{
			name:     "Исключение строки (invert match)",
			line:     "DEBUG healthcheck ok",
			query:    "",
			invert:   "healthcheck",
			expected: false,
		},
		{
			name:     "Отсев старой строки (since)",
			line:     "2023-01-01T10:00:00Z INFO old message",
			since:    timeRef, // Ожидаем логи, начиная с 12:00
			expected: false,
		},
		{
			name:     "Пропуск новой строки (since)",
			line:     "[2023-01-01 14:00:00] INFO new message",
			since:    timeRef,
			expected: true,
		},
		{
			name:     "JSON с таймстемпом не проходит until",
			line:     `{"timestamp": "2023-01-01T14:00:00Z", "msg": "future message"}`,
			until:    timeRef, // Ожидаем логи до 12:00
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Инициализируем реальный парсер с параметрами из теста
			p := NewParser(tt.level, tt.query, tt.invert, tt.since, tt.until)

			// Вызываем реальный метод.
			// Если Parse возвращает непустую строку, значит она прошла фильтры.
			parsedLine := p.Parse(tt.line)
			result := parsedLine != ""

			if result != tt.expected {
				t.Errorf("Для строки %q ожидалось %v, получили %v", tt.line, tt.expected, result)
			}
		})
	}
}
