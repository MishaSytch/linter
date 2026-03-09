# LogLint (v1.0.0)

LogLint — это специализированный линтер для Go, разработанный для контроля качества логирования. Он проверяет сообщения логгеров (`log`, `slog`, `zap` и объекта `logger`) на соответствие правилам стиля проекта.

## Возможности

### 1. Основные правила проверки:
- **Регистр**: Сообщение должно начинаться со строчной буквы (игнорируя цифры и знаки)
- **Язык**: Сообщения должны быть на английском языке (латиница)
- **Символы**: Запрещено использование спецсимволов и эмодзи (не-ASCII символы)
- **Безопасность**: Поиск утечек чувствительных данных (`password`, `token`, `api_key`, `secret`) как в тексте сообщений, так и в именах передаваемых переменных

### 2. Поддерживаемые возможности анализа:
- Анализ констант (`const`) и строковых литералов
- Обработка конкатенации строк (например, `"error: " + reason`)
- Рекурсивный анализ внутри `fmt.Sprintf`
- Проверка всех аргументов функции логирования
- Конфигурация правил, вывод подсказок, настройка формата вывода ошибок (скопом или по одной)

### 3. Ограничения анализа:
- Линтер не может анализировать строки, возвращаемые функциями в рантайме (например, `log.Print(msgFromFunc())`)
- Если в проекте используется собственная функция-обертка над логгером, он ее пропустит, так как это уже не статический анализ
  
  Пример: 
  ```go
  logErr := func(m string) {
    logger.Error(m)
  }
  logErr("Upper case error")
  ```
  Но, если в передоваемой строке или объекте есть намек на указанные в конфиге правила – линтер выведет замечание
  
  Пример из теста:
  
  -  Выводится два замечания, так как "token" встречается в двух местах
  ``` go
  func (r *Request) MarshalLogObject(enc zapcore.ObjectEncoder) error {
  	enc.AddString("id", r.ID)
  	enc.AddString("token", r.Token) // want "log message contains potentially sensitive data: token" "log message contains potentially sensitive data: token"
  	return nil
  }
  ```
  
  -  Ограничиваем только вывод, а не наличие поля в структуре
  ``` go
  _ = &Request{ID: "123", Token: "secret-abc"} // здесь должно быть нормально, так как мы не выводим никуда эту структуру
  
  req := &Request{ID: "123", Token: "secret-abc"} // want "log message contains potentially sensitive data: secret" "log message contains potentially sensitive data: token" "log message contains potentially sensitive data: secret" "log message contains potentially sensitive data: token"
  logger.Info("processing request", zap.Object("req", req))
  ```

## Запуск

### Как отдельная утилита

В проекте есть пример конфигурации и тестовый проект:
- `./internal/exampleProject/`
- `./config.yml`

Для демонстрации работы линтера выполните команду:

```bash
go run ./cmd/linter/main.go ./internal/exampleProject/...
```

### Интеграция с golangci-lint

Для работы линтера в качестве плагина необходимо выполнить сборку динамической библиотеки.


#### 1. Подготовка окружения

Убедитесь, что у вас установлена совместимая версия golangci-lint (рекомендуется v1.64.5, собранная на Go 1.24.5):

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5
```


#### 2. Сборка плагина

Для корректной работы плагина обязательно включите CGO:

для fish:

```fish
set -x CGO_ENABLED 1; go build -buildmode=plugin -o ./loglint.so ./plugin/main.go
```

для bash/zsh:

```bash
CGO_ENABLED=1 go build -buildmode=plugin -o ./loglint.so ./plugin/main.go
```

#### 3. Запуск анализа

Используйте готовую конфигурацию `.golangci.yml`

```bash
golangci-lint run ./internal/exampleProject/...
```

## Пример настройки и использования

В `./internal/exampleProject/` находится демо проект, который требуется проверить линтером. 
Данное руководство покажет несколько способов использования данного линтера.

### Детальный разбор

Линтер поддерживает режим детального разбора каждой ошибки. Для этого необходимо настроить конфигурацию `config.yml` 
в корне проекта:

```yml
loggers: # список анализируемых логеров
  - log
  - slog
  - zap
  - myCustomLogger

sensitive_rules: # правила для поиска чувствительных данных
  patterns: # паттерн поиска
    - name: "Any Card"
      regex: '[0-9]{13,16}'

  words: # чувствительные слова
    - password
    - token
    - secret


output: # способы вывода информации
  show_in_console: true # вывод в консоль
  show_suggestions: false # вывод в idea
  errors_aggregate: false # способ описания ошибок, для детального - false
  is_test: false # тесты чувствительны к выводу ошибок
```

Далее требуется запустить линтер любым удобным способом, описанным выше.

Ожидаемый вывод:
```fish
go run ./cmd/linter/main.go ./internal/exampleProject/...
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:14:14: log message should start with a lowercase letter 
        suggested:      ой! 🚀
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:14:14: log message should be in English
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:14:14: log message shouldn`t contain special characters or emojis  
        suggested:      Ой! 
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:16:14: log message should start with a lowercase letter 
        suggested:      current password: %s and token: %s
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:16:14: log message contains potentially sensitive data 
        suggested:      Current ********: %s and *****: %s
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:18:13: log message should start with a lowercase letter 
        suggested:      eRROR!!! Что-то пошло не так... ∑(O_O;)
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:18:13: log message should be in English
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:18:13: log message shouldn`t contain special characters or emojis  
        suggested:      ERROR!!! Что-то пошло не так... (O_O;)
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:23:14: log message should be in English
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:24:12: log message contains potentially sensitive data 
        suggested:      take my money credit card: ********
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/cmd/main.go:9:12: log message should start with a lowercase letter 
        suggested:      это тестовый проект для линтера логов!
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/cmd/main.go:9:12: log message should be in English
exit status 3
```


### Комплексный разбор

Линтер поддерживает режим комплексного разбора, при котором ошибки формируют общий итог и рекомендации. 
Для этого необходимо настроить конфигурацию `config.yml`в корне проекта (для более подробного разбора смотри раздел [Детальный разбор](#детальный-разбор)):

```yml
loggers:
  - log
  - slog
  - zap
  - myCustomLogger

sensitive_rules:
  patterns:
    - name: "Any Card"
      regex: '[0-9]{13,16}'

  words:
    - password
    - token
    - secret


output:
  show_in_console: true
  show_suggestions: false
  errors_aggregate: true # для агрегации - false
  is_test: false
```

Далее требуется запустить линтер любым удобным способом, описанным выше.

Ожидаемый вывод:
```fish
go run ./cmd/linter/main.go ./internal/exampleProject/...
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:14:14: log message issues:
  - should start with a lowercase letter
  - log message should be in English
  - log message shouldn`t contain special characters or emojis
        suggested:      "! "
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:16:14: log message issues:
  - should start with a lowercase letter
  - log message contains potentially sensitive data: password, token
        suggested:      "current ********: %s and *****: %s"
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:18:13: log message issues:
  - should start with a lowercase letter
  - log message should be in English
  - log message shouldn`t contain special characters or emojis
        suggested:      "eRROR!!! -   ... (O_O;)"
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:23:14: log message issues:
  - log message should be in English
        suggested:      " "
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/internal/service.go:24:12: log message issues:
  - log message contains potentially sensitive data: Any Card
        suggested:      "take my money credit card: ********"
/Users/michealsytch/GolandProjects/linter/internal/exampleProject/cmd/main.go:9:12: log message issues:
  - should start with a lowercase letter
  - log message should be in English
        suggested:      "     !"
exit status 3
```


## Тестирование

Для тестирования необходимо добавить зависимости zap в `./internal/testdata/src`, чтобы `analyzer_test.go` знал об его 
существовании (он запускается в изолированном окружении). Для этого создайте `vendor` и переместите зависимости туда:

```bash
# 1. Создаем вендор
go mod vendor
# или, если есть go.work, где подключен demo проект ./internal/exampleProject
go work vendor

 # 2. Создаем структуру папок
 mkdir -p internal/testdata/src/go.uber.org/zap
 mkdir -p internal/testdata/src/go.uber.org/multierr

 # 3. Копируем исходники
 cp -r vendor/go.uber.org/zap/. internal/testdata/src/go.uber.org/zap/
 cp -r vendor/go.uber.org/multierr/. internal/testdata/src/go.uber.org/multierr/
```

Теперь можно запустить тесты:
```bash
 go test -v ./...
```
