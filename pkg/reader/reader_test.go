package reader

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"logspector/pkg/formatter"
	"logspector/pkg/parser"

	"github.com/fatih/color"
)

// TestReader_FileRead демонстрирует паттерн тестирования чтения из файлов
func TestReader_FileRead(t *testing.T) {
	// 1. Создаем временный файл
	tmpFile, err := os.CreateTemp("", "logspector_test_*.log")
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	// Обязательно удаляем файл после завершения теста
	defer os.Remove(tmpFile.Name())

	// 2. Записываем тестовые данные, которые утилита должна прочитать
	testData := "[INFO] Test message 1\n[ERROR] Test error 2\n"
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Ошибка записи в файл: %v", err)
	}
	tmpFile.Close() // Закрываем, чтобы ваш Reader мог открыть его заново

	// 3. Инициализируем парсер и форматтер без фильтров (выключаем цвета для удобного сравнения)
	p := parser.NewParser("", "", "", time.Time{}, time.Time{})
	f := formatter.NewFormatter(false, true) // noColor = true

	// Перехватываем stdout (так же, как делали в тестах форматтера)
	oldStdout := os.Stdout
	oldColorOutput := color.Output
	rPipe, wPipe, _ := os.Pipe()
	os.Stdout = wPipe
	color.Output = wPipe

	// 4. Создаем ридер и контекст с отменой
	r := NewReader(tmpFile.Name())
	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем ReadAndProcess в отдельной горутине, так как внутри он слушает изменения файла (tail -f)
	errCh := make(chan error)
	go func() {
		errCh <- r.ReadAndProcess(ctx, p, f)
	}()

	// Даем горутине немного времени, чтобы она успела прочитать файл
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст, чтобы выйти из бесконечного цикла чтения
	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("Ошибка при чтении: %v", err)
	}

	// Восстанавливаем stdout и читаем результат
	wPipe.Close()
	os.Stdout = oldStdout
	color.Output = oldColorOutput
	var buf bytes.Buffer
	io.Copy(&buf, rPipe)

	result := strings.TrimSpace(buf.String())
	expected := "[INFO] Test message 1\n[ERROR] Test error 2"
	if result != expected {
		t.Errorf("Ожидалось:\n%q\nПолучили:\n%q", expected, result)
	}
}
