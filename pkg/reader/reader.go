package reader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bin2738/logspector/pkg/formatter"
	"github.com/bin2738/logspector/pkg/parser"

	"github.com/fsnotify/fsnotify"
)

type Reader struct {
	filePaths       []string
	dockerContainer string
	kubePod         string
	kubeArgs        []string
}

func NewReader(filePath, dockerContainer, kubePod, kubeArgsStr string) *Reader {
	var paths []string
	if filePath != "" {
		for _, p := range strings.Split(filePath, ",") {
			paths = append(paths, strings.TrimSpace(p))
		}
	}
	var kubeArgs []string
	if kubeArgsStr != "" {
		// strings.Fields правильно разбивает строку по пробелам, как это делает shell
		kubeArgs = strings.Fields(kubeArgsStr)
	}
	return &Reader{filePaths: paths, dockerContainer: dockerContainer, kubePod: kubePod, kubeArgs: kubeArgs}
}

func (r *Reader) ReadAndProcess(ctx context.Context, p *parser.Parser, f *formatter.Formatter) error {
	if r.kubePod != "" {
		return r.readFromKube(ctx, p, f)
	}

	if r.dockerContainer != "" {
		return r.readFromDocker(ctx, p, f)
	}

	if len(r.filePaths) == 0 {
		return r.readFromStdin(ctx, p, f)
	}

	errCh := make(chan error, len(r.filePaths))
	for _, fp := range r.filePaths {
		go func(path string) {
			errCh <- r.readFromFile(ctx, path, p, f)
		}(fp)
	}

	for i := 0; i < len(r.filePaths); i++ {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Reader) readFromFile(ctx context.Context, filePath string, p *parser.Parser, f *formatter.Formatter) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("ошибка создания watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(filePath)
	if err != nil {
		return fmt.Errorf("ошибка отслеживания файла: %w", err)
	}

	// Используем bufio.Reader вместо Scanner, так как Scanner плохо переносит дозапись после EOF
	reader := bufio.NewReader(file)

	var prefix string
	if len(r.filePaths) > 1 {
		prefix = filepath.Base(filePath)
	}

	// Функция для чтения доступных на данный момент строк
	readAvailableLines := func() {
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				// Убираем символ переноса строки для парсера
				cleanLine := strings.TrimSuffix(line, "\n")
				parsedLine := p.Parse(cleanLine)
				f.Format(parsedLine, prefix)
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
			f.Format(parsedLine, "")
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

func (r *Reader) readFromDocker(ctx context.Context, p *parser.Parser, f *formatter.Formatter) error {
	cmd := exec.CommandContext(ctx, "docker", "logs", "-f", r.dockerContainer)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ошибка получения stdout контейнера: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("ошибка получения stderr контейнера: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска docker logs: %w", err)
	}

	prefix := r.dockerContainer
	scanStream := func(stream io.Reader) {
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			line := scanner.Text()
			parsedLine := p.Parse(line)
			f.Format(parsedLine, prefix)
		}
	}

	// Читаем оба потока одновременно (ошибки и обычные логи)
	go scanStream(stdout)
	go scanStream(stderr)

	// Ждем завершения команды. Если контекст отменен (Ctrl+C), Wait завершится
	cmd.Wait()
	return nil
}

func (r *Reader) readFromKube(ctx context.Context, p *parser.Parser, f *formatter.Formatter) error {
	// Динамически собираем аргументы для kubectl
	args := []string{"logs", "-f", r.kubePod}
	args = append(args, r.kubeArgs...)

	cmd := exec.CommandContext(ctx, "kubectl", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ошибка получения stdout pod'а: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("ошибка получения stderr pod'а: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска kubectl logs: %w", err)
	}

	prefix := r.kubePod
	scanStream := func(stream io.Reader) {
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			line := scanner.Text()
			parsedLine := p.Parse(line)
			f.Format(parsedLine, prefix)
		}
	}

	// Читаем оба потока одновременно (ошибки и обычные логи)
	go scanStream(stdout)
	go scanStream(stderr)

	// Ждем завершения команды
	cmd.Wait()
	return nil
}
