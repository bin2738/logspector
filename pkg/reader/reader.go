package reader

import (
	"bufio"
	"fmt"
	"os"

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

func (r *Reader) ReadAndProcess(p *parser.Parser, f *formatter.Formatter) {
	if r.filePath != "" {
		r.readFromFile(p, f)
	} else {
		r.readFromStdin(p, f)
	}
}

func (r *Reader) readFromFile(p *parser.Parser, f *formatter.Formatter) {
	file, err := os.Open(r.filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Читаем всё содержимое файла сначала
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parsedLine := p.Parse(line)
		f.Format(parsedLine)
	}

	// Теперь слежение за файлом
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add(r.filePath)
	if err != nil {
		panic(err)
	}

	// Читаем новые строки по событиям fsnotify
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Открываем файл заново для чтения новых строк
				file, err = os.Open(r.filePath)
				if err != nil {
					fmt.Println("Error reopening file:", err)
					continue
				}

				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					parsedLine := p.Parse(line)
					f.Format(parsedLine)
				}
				file.Close()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("watcher error:", err)
		}
	}
}

func (r *Reader) readFromStdin(p *parser.Parser, f *formatter.Formatter) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parsedLine := p.Parse(line)
		f.Format(parsedLine)
	}
}
