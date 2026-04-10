# Logspector

**Logspector** — консольная утилита на Go для удобного просмотра и анализа текстовых логов в реальном времени с возможностью цветовой подсветки и фильтрации.

---

## 📋 Описание

Logspector позволяет:
- Читать логи из файла в режиме `tail -f` (отслеживание изменений).
- Фильтровать логи по уровню (`ERROR`, `WARN`, `INFO`, `DEBUG`).
- Подсвечивать строки цветом:
  - `ERROR` — красный
  - `WARN` — жёлтый
  - `INFO` — зеленый
  - `DEBUG` — синий
- Pretty print JSON-строк (если включён флаг `-json`).
- Читать логи из stdin.

---

## 🧰 Требования

- Go 1.22+

---

## 🚀 Установка

```bash
go build -o logspector cmd/logspector/main.go
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