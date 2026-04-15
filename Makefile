APP_NAME = logspector
# Версия по умолчанию. Можно переопределить при запуске: make build VERSION=2.0.0
# Пытаемся взять версию из git-тега, если не получается - ставим 1.0.1
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "1.0.1")
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

.PHONY: all build build-linux build-windows build-macos build-all test cover clean run deps

all: test build

build:
	@echo "🚀 Собираем $(APP_NAME) версии $(VERSION)..."
	go build $(LDFLAGS) -o $(APP_NAME) ./cmd/logspector/main.go

build-linux:
	@echo "🐧 Собираем для Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-linux ./cmd/logspector/main.go

build-windows:
	@echo "🪟 Собираем для Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME).exe ./cmd/logspector/main.go

build-macos:
	@echo "🍎 Собираем для macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-macos ./cmd/logspector/main.go

build-all: build-linux build-windows build-macos

test:
	@echo "🧪 Запускаем тесты..."
	go test -v ./...

cover:
	@echo "📊 Проверка покрытия кода тестами..."
	go test -cover ./...

run: build
	@echo "▶️ Запускаем $(APP_NAME)..."
	./$(APP_NAME)

clean:
	@echo "🧹 Очистка..."
	rm -f $(APP_NAME)
	rm -f $(APP_NAME)-linux $(APP_NAME).exe $(APP_NAME)-macos

deps:
	@echo "📦 Обновление зависимостей..."
	go mod tidy