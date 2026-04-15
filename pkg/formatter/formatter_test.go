package formatter

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestFormatter_Colors(t *testing.T) {
	// Принудительно включаем цвета для тестов (иначе пакет отключит их при записи не в терминал, а в pipe)
	color.NoColor = false

	tests := []struct {
		name     string
		level    string
		noColor  bool
		expected string
	}{
		{
			name:     "ERROR с цветом",
			level:    "ERROR",
			noColor:  false,
			expected: "\033[31mERROR\033[0m", // ANSI-код для красного цвета
		},
		{
			name:     "ERROR без цвета (-no-color)",
			level:    "ERROR",
			noColor:  true,
			expected: "ERROR", // Ожидаем чистый текст без спецсимволов
		},
		{
			name:     "INFO без цвета",
			level:    "INFO",
			noColor:  true,
			expected: "INFO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Инициализируем форматтер: JSON выключен, noColor берем из теста
			f := NewFormatter(false, tt.noColor)

			// Перехватываем стандартный вывод (stdout), так как Format делает fmt.Println
			oldStdout := os.Stdout
			oldColorOutput := color.Output // Сохраняем оригинальный вывод fatih/color
			r, w, _ := os.Pipe()
			os.Stdout = w
			color.Output = w // Перенаправляем цветной вывод в наш pipe

			f.Format(tt.level, "")

			// Восстанавливаем stdout и читаем результат
			w.Close()
			os.Stdout = oldStdout
			color.Output = oldColorOutput
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Убираем перенос строки, который добавляет Println
			result := strings.TrimSpace(buf.String())

			if result != tt.expected {
				t.Errorf("Ожидалось %q, получили %q", tt.expected, result)
			}
		})
	}
}
