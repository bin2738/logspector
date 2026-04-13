package reader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"logspector/pkg/formatter"
	"logspector/pkg/parser"

	"github.com/fsnotify/fsnotify"
)

type Reader struct {
	filePath string
}

func NewReader(filePath string) *Reader {
	return &Reader{filePath: filePath}
}

func (r *Reader) ReadAndProcess(ctx context.Context, p *parser.Parser, f *formatter.Formatter) error {
	if r.filePath != "" {
		return r.readFromFile(ctx, p, f)
	}
	return r.readFromStdin(ctx, p, f)
}

func (r *Reader) readFromFile(ctx context.Context, p *parser.Parser, f *formatter.Formatter) error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("ошибка создания watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(r.filePath)
	if err != nil {
		return fmt.Errorf("ошибка отслеживания файла: %w", err)
	}

	// Используем bufio.Reader вместо Scanner, так как Scanner плохо переносит дозапись после EOF
	reader := bufio.NewReader(file)

	// Функция для чтения доступных на данный момент строк
	readAvailableLines := func() {
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				// Убираем символ переноса строки для парсера
				cleanLine := strings.TrimSuffix(line, "\n")
				parsedLine := p.Parse(cleanLine)
				f.Format(parsedLine)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "ошибка чтения: %v\n", err)
				break
			}
		}
	}

	// Сначала читаем всё содержимое файла, которое есть на данный момент
	readAvailableLines()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Файл дозаписан, просто продолжаем читать открытый файл с текущего места
				readAvailableLines()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "ошибка watcher: %v\n", err)
		}
	}
}

func (r *Reader) readFromStdin(ctx context.Context, p *parser.Parser, f *formatter.Formatter) error {
	scanner := bufio.NewScanner(os.Stdin)
	errCh := make(chan error, 1)

	// Запускаем чтение в горутине, чтобы иметь возможность прервать работу по контексту,
	// так как scanner.Scan() блокирует поток до появления новых данных.
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			parsedLine := p.Parse(line)
			f.Format(parsedLine)
		}
		errCh <- scanner.Err()
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}
