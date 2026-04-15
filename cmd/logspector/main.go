package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"github.com/bin2738/logspector/pkg/formatter"
	"github.com/bin2738/logspector/pkg/parser"
	"github.com/bin2738/logspector/pkg/reader"
)

type Config struct {
	Files    string `yaml:"files"`
	Level    string `yaml:"level"`
	JSON     bool   `yaml:"json"`
	Query    string `yaml:"query"`
	Invert   string `yaml:"invert"`
	NoColor  bool   `yaml:"no_color"`
	Since    string `yaml:"since"`
	Until    string `yaml:"until"`
	Docker   string `yaml:"docker"`
	Kube     string `yaml:"kube"`
	KubeArgs string `yaml:"kube_args"`
}

// Version содержит текущую версию утилиты.
// Может быть перезаписана при сборке: go build -ldflags "-X main.Version=0.1.0"
var Version = "0.1.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("Logspector version %s\n", Version)
		os.Exit(0)
	}

	configFlag := flag.String("c", "", "Путь к файлу конфигурации (.yaml)")
	flag.String("f", "", "Path to log file")
	flag.String("l", "", "Log level to filter (ERROR, WARN, INFO, DEBUG)")
	flag.Bool("json", false, "Pretty print JSON")
	flag.String("q", "", "Поиск по подстроке (grep)")
	flag.String("v", "", "Исключить строки с текстом (invert match)")
	flag.Bool("no-color", false, "Отключить цветовую подсветку")
	flag.String("since", "", "Показывать логи начиная с этого времени (например: 5m, 1h)")
	flag.String("until", "", "Показывать логи до этого времени (например: 5m, 1h)")
	flag.String("docker", "", "Имя или ID Docker контейнера для чтения логов")
	flag.String("kube", "", "Имя Kubernetes pod'а для чтения логов")
	flag.String("kube-args", "", "Дополнительные аргументы для kubectl (например, \"-n my-namespace\")")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Logspector - консольная утилита для просмотра логов в реальном времени.
Использование:
  %s [флаги]

Флаги:
`, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Примеры:
  ./logspector version                        # Показать версию утилиты
  ./logspector -c config.yaml                 # Использовать файл конфигурации
  ./logspector -f app.log                     # Смотреть логи в реальном времени
  ./logspector -f app.log -l ERROR            # Показывать только ошибки
  ./logspector -f app.log -q "user_id" -v "ok" # Поиск по тексту и исключение
  ./logspector -f app.log -no-color           # Вывод без цветовой подсветки
  ./logspector -f app.log -since 5m           # Вывод логов только за последние 5 минут
  ./logspector -docker my-container           # Читать логи из Docker-контейнера
  ./logspector -kube my-pod                   # Читать логи из Kubernetes pod'а
  ./logspector -kube my-pod -kube-args "-n staging" # Читать логи из pod'а в другом namespace
  cat old.log | ./logspector -json            # Обработать JSON из stdin`)
	}

	flag.Parse()

	var cfg Config

	// Определяем путь к конфигу: либо из флага, либо ищем стандартные
	configPath := *configFlag
	if configPath == "" {
		if _, err := os.Stat(".logspector.yaml"); err == nil {
			configPath = ".logspector.yaml"
		} else if home, err := os.UserHomeDir(); err == nil {
			defaultHome := filepath.Join(home, ".logspector.yaml")
			if _, err := os.Stat(defaultHome); err == nil {
				configPath = defaultHome
			}
		}
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка парсинга конфига %s: %v\n", configPath, err)
			}
		} else {
			if *configFlag != "" { // Выводим ошибку только если конфиг был передан явно
				fmt.Fprintf(os.Stderr, "Ошибка чтения конфига %s: %v\n", configPath, err)
			}
		}
	}

	// Функция для применения флагов поверх конфигурации
	applyFlags := func(c *Config) {
		flag.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "f":
				c.Files = f.Value.String()
			case "l":
				c.Level = f.Value.String()
			case "json":
				c.JSON = f.Value.String() == "true"
			case "q":
				c.Query = f.Value.String()
			case "v":
				c.Invert = f.Value.String()
			case "no-color":
				c.NoColor = f.Value.String() == "true"
			case "since":
				c.Since = f.Value.String()
			case "until":
				c.Until = f.Value.String()
			case "docker":
				c.Docker = f.Value.String()
			case "kube":
				c.Kube = f.Value.String()
			case "kube-args":
				c.KubeArgs = f.Value.String()
			}
		})
	}

	applyFlags(&cfg)

	since := parseTimeFlag(cfg.Since)
	until := parseTimeFlag(cfg.Until)

	// Создаем парсер и форматтер
	p := parser.NewParser(cfg.Level, cfg.Query, cfg.Invert, since, until)
	f := formatter.NewFormatter(cfg.JSON, cfg.NoColor)

	// Контекст для плавного завершения (Graceful Shutdown)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Отслеживание изменений файла конфигурации в реальном времени
	if configPath != "" {
		watcher, err := fsnotify.NewWatcher()
		if err == nil {
			defer watcher.Close()
			watcher.Add(configPath)
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case event, ok := <-watcher.Events:
						if !ok {
							return
						}
						if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
							time.Sleep(100 * time.Millisecond) // Ждем завершения записи в файл редактором
							var newCfg Config
							if data, err := os.ReadFile(configPath); err == nil {
								if err := yaml.Unmarshal(data, &newCfg); err == nil {
									applyFlags(&newCfg)
									p.Update(newCfg.Level, newCfg.Query, newCfg.Invert, parseTimeFlag(newCfg.Since), parseTimeFlag(newCfg.Until))
									f.Update(newCfg.JSON, newCfg.NoColor)
									fmt.Fprintf(os.Stderr, "🔄 Конфигурация %s перезагружена!\n", configPath)
								}
							}
						}
					}
				}
			}()
		}
	}

	// Запускаем чтение (если файл не указан, cfg.Files будет "", и ридер переключится на stdin)
	r := reader.NewReader(cfg.Files, cfg.Docker, cfg.Kube, cfg.KubeArgs)
	if err := r.ReadAndProcess(ctx, p, f); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при чтении: %v\n", err)
		os.Exit(1)
	}
}

func parseTimeFlag(val string) time.Time {
	if val == "" {
		return time.Time{}
	}
	// Пытаемся распарсить как длительность (например "5m", "1h")
	if d, err := time.ParseDuration(val); err == nil {
		return time.Now().Add(-d) // Относительное время в прошлом
	}
	// Пытаемся распарсить как RFC3339 (например "2023-10-12T07:20:50.52Z")
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t
	}
	fmt.Fprintf(os.Stderr, "Внимание: не удалось распарсить время '%s'. Используйте длительность (5m) или RFC3339.\n", val)
	return time.Time{}
}
