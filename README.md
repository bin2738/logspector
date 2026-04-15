# 🔎 Logspector

[![Go Report Card](https://goreportcard.com/badge/github.com/bin2738/logspector)](https://goreportcard.com/report/github.com/bin2738/logspector)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Logspector** — это легковесная консольная утилита на Go для удобного просмотра и анализа логов в реальном времени. Она идеально подходит для локальной разработки и отладки, предоставляя мощные возможности фильтрации и подсветки.

---

## ✨ Возможности

Logspector позволяет:
- **Чтение и отслеживание**: Мониторинг одного или **нескольких файлов** одновременно (в режиме `tail -f`) с умным добавлением префиксов файлов.
- **Продвинутая фильтрация**:
    - По уровню (`ERROR`, `WARN`, `INFO`, `DEBUG`).
    - По тексту с помощью `-q` (как `grep`).
    - Исключение строк с помощью `-v` (как `grep -v`).
    - По временным меткам с флагами `-since` и `-until` (поддержка `5m`, `1h` или RFC3339).
- **Умная подсветка**:
    - Автоматическая подсветка уровней логов (`ERROR` — красный, `WARN` — жёлтый и т.д.).
    - Подсветка имени файла при чтении нескольких источников.
    - Отключение цвета флагом `-no-color` для удобного перенаправления вывода в файлы.
- **Работа с JSON**: Автоматическое извлечение, определение и форматирование (`pretty-print`) JSON-строк.
- **Гибкая конфигурация**:
    - Поддержка файла конфигурации `.logspector.yaml` для задания настроек по умолчанию.
    - **Горячая перезагрузка** (Hot Reload) конфигурации "на лету" без перезапуска утилиты.
    - Приоритет флагов командной строки над настройками из файла.
- **Кросс-платформенность**: Работает на Linux, macOS и Windows.
- **Чтение из Stdin**: Возможность обрабатывать логи, переданные через `pipe`.

---

## 🚀 Установка

### С помощью `go install` (рекомендуется)
```bash
go install github.com/bin2738/logspector/cmd/logspector@latest
```

### Из исходного кода
```bash
git clone https://github.com/bin2738/logspector.git
cd logspector
make build
```

---

## 📖 Использование

### Просмотр логов в реальном времени

```bash
./logspector -f /var/log/myapp.log
```

### Фильтрация по уровню

```bash
./logspector -f /var/log/myapp.log -l ERROR
```

### Pretty print JSON

```bash
cat old.log | ./logspector -json
```

### Показать справку

```bash
./logspector -h
```

---

## 📦 Флаги

| Флаг     | Описание                                       |
|----------|------------------------------------------------|
| `-c`     | Путь к файлу конфигурации (.yaml)              |
| `-f`     | Путь к файлу с логами (можно указать несколько через запятую) |
| `-l`     | Фильтр по уровню лога (`ERROR`, `WARN`, `INFO`, `DEBUG`) |
| `-q`     | Поиск по подстроке (grep)                      |
| `-v`     | Исключить строки, содержащие текст (invert match) |
| `-since` | Показывать логи начиная с этого времени (например: `5m`) |
| `-until` | Показывать логи до этого времени               |
| `-no-color`| Отключить цветовую подсветку                 |
| `-json`  | Принудительный pretty print для JSON           |
| `-h`     | Показать справку                               |

---

## 🪄 Автодополнение (Autocomplete)

В проекте есть скрипт для поддержки автодополнения флагов и путей к файлам по нажатию клавиши `Tab`.

**Для Bash:**
Добавьте загрузку скрипта в ваш `~/.bashrc`:
```bash
source /путь/к/logspector/autocomplete.sh
```

**Для Zsh:**
Zsh поддерживает скрипты от bash через модуль совместимости. Добавьте в `~/.zshrc`:
```zsh
autoload -Uz bashcompinit && bashcompinit
source /путь/к/logspector/autocomplete.sh
```

После этого перезапустите терминал. Теперь при вводе `logspector -<TAB>` вы увидите список доступных флагов, а для `logspector -l <TAB>` — варианты `INFO`, `ERROR` и т.д.

---

## 📁 Структура проекта

```
logspector/
├── cmd/
│   └── logspector/
│       └── main.go
├── pkg/
│   ├── reader/     # Чтение и отслеживание файлов
│   ├── parser/     # Парсинг уровней логов и JSON
│   └── formatter/  # Цветной вывод и pretty print
├── go.mod
└── README.md
```

---

## 🧪 Примеры

```bash
# Смотреть логи в реальном времени
./logspector -f app.log

# Показывать только ошибки
./logspector -f app.log -l ERROR

# Поиск по тексту и исключение шума
./logspector -f app.log -q "user_id=123" -v "healthcheck"

# Чтение нескольких файлов без цвета (удобно для перенаправления в файл)
./logspector -f app.log,db.log -no-color > combined.log

# Вывод логов только за последние 5 минут
./logspector -f app.log -since 5m

# Обработать JSON из stdin
cat log.json | ./logspector -json
```

---
# Сборка для Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o logspector-linux ./cmd/logspector

# Сборка для Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o logspector.exe ./cmd/logspector

---

## 📝 Лицензия

MIT